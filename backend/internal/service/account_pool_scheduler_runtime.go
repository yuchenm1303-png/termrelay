package service

import "time"

// AccountPoolScheduler is the provider-neutral runtime facade used by gateway
// orchestration. It deliberately owns only ranking feedback and the transient
// circuit breaker; hard eligibility checks, concurrency acquisition, waiting,
// failover and billing remain in the existing gateway lifecycle.
type AccountPoolScheduler struct {
	stats   *accountPoolRuntimeStats
	breaker *accountPoolCircuitBreaker
	weights AccountPoolScoreWeights
	topK    int
}

// NewAccountPoolScheduler creates an isolated scheduler runtime. A non-positive
// topK means all scored candidates remain in the rendezvous set.
func NewAccountPoolScheduler(topK int, weights AccountPoolScoreWeights) *AccountPoolScheduler {
	if weights == (AccountPoolScoreWeights{}) {
		weights = defaultAccountPoolScoreWeights()
	}
	return &AccountPoolScheduler{
		stats:   newAccountPoolRuntimeStats(),
		breaker: &accountPoolCircuitBreaker{},
		weights: weights,
		topK:    topK,
	}
}

// Rank scores candidates that have already passed the gateway's hard filters.
// It has no side effects: in particular it does not claim a half-open circuit
// probe. Call Allow immediately before attempting to acquire/send on a ranked
// account.
func (s *AccountPoolScheduler) Rank(inputs []accountPoolCandidateInput, seed string, now time.Time) []accountPoolCandidateScore {
	if s == nil || len(inputs) == 0 {
		return nil
	}
	if now.IsZero() {
		now = time.Now()
	}
	stats := s.stats
	if stats == nil {
		stats = newAccountPoolRuntimeStats()
		s.stats = stats
	}
	weights := s.weights
	if weights == (AccountPoolScoreWeights{}) {
		weights = defaultAccountPoolScoreWeights()
	}
	return rankAccountPoolCandidates(
		scoreAccountPoolCandidates(inputs, stats, weights, now),
		s.topK,
		seed,
	)
}

// Allow checks the fast transient circuit immediately before the caller tries
// to use an account. When an open circuit reaches its cooldown deadline this
// call atomically claims the single half-open probe slot.
func (s *AccountPoolScheduler) Allow(accountID int64, now time.Time) bool {
	if s == nil || accountID <= 0 {
		return false
	}
	if now.IsZero() {
		now = time.Now()
	}
	if s.breaker == nil {
		s.breaker = &accountPoolCircuitBreaker{}
	}
	return s.breaker.allow(accountID, now)
}

// Report records the result of one actual upstream attempt. transient must only
// be true for temporary transport/upstream failures; authentication, permanent
// request errors and persistent provider rate-limit state stay owned by the
// existing gateway/account state machine.
func (s *AccountPoolScheduler) Report(accountID int64, success, transient bool, firstTokenMs *int, now time.Time) {
	if s == nil || accountID <= 0 {
		return
	}
	if now.IsZero() {
		now = time.Now()
	}
	if s.stats == nil {
		s.stats = newAccountPoolRuntimeStats()
	}
	if s.breaker == nil {
		s.breaker = &accountPoolCircuitBreaker{}
	}
	s.stats.report(accountID, success, firstTokenMs)
	s.breaker.report(accountID, success, transient, now)
}

// Feedback returns scheduler-only health telemetry. It never exposes request
// bodies, model payloads, API keys, OAuth tokens or other credentials.
func (s *AccountPoolScheduler) Feedback(accountID int64) accountPoolFeedback {
	if s == nil || s.stats == nil {
		return accountPoolFeedback{}
	}
	return s.stats.snapshot(accountID)
}
