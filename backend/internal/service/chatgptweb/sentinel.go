package chatgptweb

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type ProofRequirement struct {
	Difficulty string `json:"difficulty,omitempty"`
	Required   bool   `json:"required"`
	Seed       string `json:"seed,omitempty"`
}
type SentinelPrepareResponse struct {
	Persona      string           `json:"persona,omitempty"`
	Token        string           `json:"token,omitempty"`
	PrepareToken string           `json:"prepare_token,omitempty"`
	Proof        ProofRequirement `json:"proofofwork"`
	Turnstile    struct {
		Required bool   `json:"required"`
		DX       string `json:"dx,omitempty"`
	} `json:"turnstile"`
	ForceLogin bool `json:"force_login"`
}
type sentinelFinalizeResponse struct {
	Token string `json:"token"`
}
type SentinelTokens struct {
	PrepareToken          string
	ChatRequirementsToken string
	ProofToken            string
}

type RequirementsTokenProvider interface {
	RequirementsToken(context.Context, *ClientState) (string, error)
}
type RequirementsTokenProviderFunc func(context.Context, *ClientState) (string, error)

func (f RequirementsTokenProviderFunc) RequirementsToken(ctx context.Context, s *ClientState) (string, error) {
	return f(ctx, s)
}

type ProofOfWorkSolver interface {
	Solve(context.Context, ProofRequirement, *ClientState) (string, error)
}
type ProofOfWorkSolverFunc func(context.Context, ProofRequirement, *ClientState) (string, error)

func (f ProofOfWorkSolverFunc) Solve(ctx context.Context, p ProofRequirement, s *ClientState) (string, error) {
	return f(ctx, p, s)
}

type SentinelFlow struct {
	client       *Client
	requirements RequirementsTokenProvider
	proofSolver  ProofOfWorkSolver
}

func NewSentinelFlow(client *Client, requirements RequirementsTokenProvider, proofSolver ProofOfWorkSolver) (*SentinelFlow, error) {
	if client == nil {
		return nil, errors.New("chatgptweb: sentinel client is required")
	}
	if requirements == nil {
		return nil, errors.New("chatgptweb: requirements token provider is required")
	}
	return &SentinelFlow{client: client, requirements: requirements, proofSolver: proofSolver}, nil
}

func (s *SentinelFlow) Run(ctx context.Context, state *ClientState) (*SentinelTokens, error) {
	requirementsToken, err := s.requirements.RequirementsToken(ctx, state)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(requirementsToken) == "" {
		return nil, errors.New("chatgptweb: empty requirements token")
	}
	resp, err := s.client.doJSON(ctx, http.MethodPost, "/sentinel/chat-requirements/prepare", map[string]string{"p": requirementsToken}, state, nil)
	if err != nil {
		return nil, err
	}
	var prepare SentinelPrepareResponse
	decodeErr := json.NewDecoder(resp.Body).Decode(&prepare)
	_ = resp.Body.Close()
	if decodeErr != nil {
		return nil, decodeErr
	}
	if prepare.ForceLogin {
		return nil, &UpstreamError{Kind: ErrorKindAuthentication, StatusCode: http.StatusUnauthorized, Message: "upstream requires login", RequiresReauth: true}
	}
	if prepare.Turnstile.Required || strings.TrimSpace(prepare.Turnstile.DX) != "" {
		return nil, &UpstreamError{Kind: ErrorKindChallenge, StatusCode: http.StatusForbidden, Message: "upstream requires interactive verification", ChallengeRequired: true}
	}
	if strings.TrimSpace(prepare.PrepareToken) == "" {
		return nil, &UpstreamError{Kind: ErrorKindProtocol, Message: "sentinel prepare token is missing"}
	}
	proofToken := ""
	if prepare.Proof.Required {
		if s.proofSolver == nil {
			return nil, &UpstreamError{Kind: ErrorKindProtocol, Message: "proof of work is required but no solver is configured"}
		}
		proofToken, err = s.proofSolver.Solve(ctx, prepare.Proof, state)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(proofToken) == "" {
			return nil, &UpstreamError{Kind: ErrorKindProtocol, Message: "proof of work solver returned an empty token"}
		}
	}
	payload := map[string]string{"prepare_token": prepare.PrepareToken}
	if proofToken != "" {
		payload["proofofwork"] = proofToken
	}
	resp, err = s.client.doJSON(ctx, http.MethodPost, "/sentinel/chat-requirements/finalize", payload, state, nil)
	if err != nil {
		return nil, err
	}
	var finalized sentinelFinalizeResponse
	decodeErr = json.NewDecoder(resp.Body).Decode(&finalized)
	_ = resp.Body.Close()
	if decodeErr != nil {
		return nil, decodeErr
	}
	if strings.TrimSpace(finalized.Token) == "" {
		return nil, &UpstreamError{Kind: ErrorKindProtocol, Message: "sentinel finalize token is missing"}
	}
	return &SentinelTokens{PrepareToken: prepare.PrepareToken, ChatRequirementsToken: finalized.Token, ProofToken: proofToken}, nil
}
