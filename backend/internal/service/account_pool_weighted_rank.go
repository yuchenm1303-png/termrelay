package service

import "sort"

// rankAccountPoolCandidatesWithStaticWeight keeps health/priority scoring as
// the gate into Top-K, then applies the operator traffic weight only inside that
// healthy set. This prevents a huge static weight from overpowering a degraded
// account's error/load penalties while still providing predictable traffic
// sharing among viable peers.
func rankAccountPoolCandidatesWithStaticWeight(candidates []accountPoolCandidateScore, topK int, seed string) []accountPoolCandidateScore {
	if len(candidates) == 0 {
		return nil
	}
	ranked := append([]accountPoolCandidateScore(nil), candidates...)
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].Score != ranked[j].Score {
			return ranked[i].Score > ranked[j].Score
		}
		if ranked[i].Account.Priority != ranked[j].Account.Priority {
			return ranked[i].Account.Priority < ranked[j].Account.Priority
		}
		return ranked[i].Account.ID < ranked[j].Account.ID
	})
	if topK <= 0 || topK > len(ranked) {
		topK = len(ranked)
	}

	top := append([]accountPoolCandidateScore(nil), ranked[:topK]...)
	sort.SliceStable(top, func(i, j int) bool {
		return accountPoolWeightedRendezvous(seed, top[i]) > accountPoolWeightedRendezvous(seed, top[j])
	})
	return append(top, ranked[topK:]...)
}

func accountPoolWeightedRendezvous(seed string, candidate accountPoolCandidateScore) float64 {
	base := accountPoolRendezvous(seed, candidate)
	if candidate.Account == nil {
		return base
	}
	return base * candidate.Account.SchedulingWeight()
}
