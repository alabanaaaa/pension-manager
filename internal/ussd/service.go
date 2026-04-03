package ussd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// USSDSession represents a USSD session state
type USSDSession struct {
	SessionID   string
	PhoneNumber string
	ServiceCode string
	State       string // "initial", "voting", "confirming", "completed"
	ElectionID  string
	CandidateID string
	Step        int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// USSDResponse represents the response to send back to Africa's Talking
type USSDResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Provider is the interface for USSD providers
type Provider interface {
	SendResponse(w http.ResponseWriter, sessionID, message string)
	HandleCallback(w http.ResponseWriter, r *http.Request) (*USSDRequest, error)
	Name() string
}

// USSDRequest represents the callback from Africa's Talking
type USSDRequest struct {
	SessionID   string `json:"sessionId"`
	ServiceCode string `json:"serviceCode"`
	PhoneNumber string `json:"phoneNumber"`
	Text        string `json:"text"`
}

// AfricaTalkingProvider implements Provider for Africa's Talking USSD
type AfricaTalkingProvider struct {
	APIKey      string
	ShortCode   string
	Environment string
}

// NewAfricaTalkingProvider creates a new Africa's Talking USSD provider
func NewAfricaTalkingProvider(apiKey, shortCode, environment string) *AfricaTalkingProvider {
	return &AfricaTalkingProvider{
		APIKey:      apiKey,
		ShortCode:   shortCode,
		Environment: environment,
	}
}

func (p *AfricaTalkingProvider) Name() string {
	return "africastalking"
}

// SendResponse sends a USSD response back to Africa's Talking
func (p *AfricaTalkingProvider) SendResponse(w http.ResponseWriter, sessionID, message string) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(message))
}

// HandleCallback parses the USSD callback from Africa's Talking
func (p *AfricaTalkingProvider) HandleCallback(w http.ResponseWriter, r *http.Request) (*USSDRequest, error) {
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("parse form: %w", err)
	}

	req := &USSDRequest{
		SessionID:   r.FormValue("sessionId"),
		ServiceCode: r.FormValue("serviceCode"),
		PhoneNumber: r.FormValue("phoneNumber"),
		Text:        r.FormValue("text"),
	}

	if req.SessionID == "" || req.PhoneNumber == "" {
		return nil, fmt.Errorf("missing required fields")
	}

	return req, nil
}

// Service manages USSD operations
type Service struct {
	provider Provider
	voting   VotingService
	sessions map[string]*USSDSession
}

// VotingService is the interface for voting operations (adapted from voting.Service)
type VotingService interface {
	GetElectionByPhone(ctx context.Context, phone string) (string, error)
	GetCandidatesForElection(ctx context.Context, electionID string) ([]Candidate, error)
	CastVote(ctx context.Context, electionID, memberID, candidateID, method string) error
	HasVoted(ctx context.Context, electionID, memberID string) (bool, error)
}

// Candidate represents a voting candidate
type Candidate struct {
	ID   string
	Name string
}

// NewService creates a new USSD service
func NewService(provider Provider, voting VotingService) *Service {
	return &Service{
		provider: provider,
		voting:   voting,
		sessions: make(map[string]*USSDSession),
	}
}

// HandleUSSD handles incoming USSD callbacks
func (s *Service) HandleUSSD(w http.ResponseWriter, r *http.Request) {
	req, err := s.provider.HandleCallback(w, r)
	if err != nil {
		slog.Error("USSD callback error", "error", err)
		s.provider.SendResponse(w, "", "CON An error occurred. Please try again.")
		return
	}

	slog.Info("USSD request received",
		"session_id", req.SessionID,
		"phone", req.PhoneNumber,
		"text", req.Text,
	)

	// Get or create session
	session := s.getOrCreateSession(req)

	// Process the USSD menu
	response := s.processMenu(session, req.Text)
	s.provider.SendResponse(w, req.SessionID, response)
}

func (s *Service) getOrCreateSession(req *USSDRequest) *USSDSession {
	session, exists := s.sessions[req.SessionID]
	if !exists {
		session = &USSDSession{
			SessionID:   req.SessionID,
			PhoneNumber: req.PhoneNumber,
			ServiceCode: req.ServiceCode,
			State:       "initial",
			Step:        0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		s.sessions[req.SessionID] = session
	} else {
		session.UpdatedAt = time.Now()
	}
	return session
}

func (s *Service) processMenu(session *USSDSession, text string) string {
	// Parse the text to determine current step
	// Africa's Talking sends cumulative text: "1", "1.2", "1.2.3", etc.
	parts := strings.Split(text, ".")
	step := len(parts)

	switch session.State {
	case "initial":
		return s.handleInitialState(session, text, step)
	case "voting":
		return s.handleVotingState(session, text, step)
	case "confirming":
		return s.handleConfirmingState(session, text, step)
	case "completed":
		return s.handleCompletedState(session, text, step)
	default:
		return "CON An error occurred. Please try again."
	}
}

func (s *Service) handleInitialState(session *USSDSession, text string, step int) string {
	if step == 0 || text == "" {
		// Initial menu
		return "CON Welcome to Pension Voting System\n1. Vote for Trustee Election\n2. Check Voting Status\n3. Help"
	}

	parts := strings.Split(text, ".")
	choice := parts[len(parts)-1]

	switch choice {
	case "1":
		// Check if member has an active election
		electionID, err := s.voting.GetElectionByPhone(context.Background(), session.PhoneNumber)
		if err != nil || electionID == "" {
			session.State = "completed"
			return "END You don't have any active elections to vote in. Contact your scheme administrator."
		}

		session.ElectionID = electionID
		session.State = "voting"
		session.Step = 1

		// Get candidates
		candidates, err := s.voting.GetCandidatesForElection(context.Background(), electionID)
		if err != nil || len(candidates) == 0 {
			session.State = "completed"
			return "END No candidates available for this election."
		}

		// Build candidate menu
		menu := "CON Select your candidate:\n"
		for i, c := range candidates {
			menu += fmt.Sprintf("%d. %s\n", i+1, c.Name)
		}
		return menu

	case "2":
		// Check voting status
		hasVoted, err := s.voting.HasVoted(context.Background(), session.ElectionID, session.PhoneNumber)
		if err != nil {
			return "END Unable to check voting status. Please try again."
		}

		session.State = "completed"
		if hasVoted {
			return "END You have already voted in this election. Thank you for participating!"
		}
		return "END You have not voted yet. Select option 1 from the main menu to vote."

	case "3":
		session.State = "completed"
		return "END Pension Voting System Help:\n1. Select 'Vote' to cast your vote\n2. Select 'Status' to check if you've voted\n3. You can only vote once per election\n4. Voting is secure and anonymous\n\nFor support, contact your scheme administrator."

	default:
		return "CON Invalid option. Please select 1, 2, or 3."
	}
}

func (s *Service) handleVotingState(session *USSDSession, text string, step int) string {
	parts := strings.Split(text, ".")
	choice := parts[len(parts)-1]

	// Get candidates for this election
	candidates, err := s.voting.GetCandidatesForElection(context.Background(), session.ElectionID)
	if err != nil || len(candidates) == 0 {
		session.State = "completed"
		return "END No candidates available."
	}

	// Parse candidate selection
	candidateIdx := 0
	fmt.Sscanf(choice, "%d", &candidateIdx)

	if candidateIdx < 1 || candidateIdx > len(candidates) {
		return "CON Invalid candidate number. Please select a valid candidate."
	}

	selectedCandidate := candidates[candidateIdx-1]
	session.CandidateID = selectedCandidate.ID
	session.State = "confirming"

	return fmt.Sprintf("CON You selected: %s\n\n1. Confirm Vote\n2. Cancel", selectedCandidate.Name)
}

func (s *Service) handleConfirmingState(session *USSDSession, text string, step int) string {
	parts := strings.Split(text, ".")
	choice := parts[len(parts)-1]

	if choice == "1" {
		// Cast the vote
		err := s.voting.CastVote(context.Background(), session.ElectionID, session.PhoneNumber, session.CandidateID, "ussd")
		if err != nil {
			session.State = "completed"
			return fmt.Sprintf("END Vote failed: %s. Please try again or contact support.", err.Error())
		}

		session.State = "completed"
		return "END Your vote has been recorded successfully! Thank you for participating in the election."
	} else if choice == "2" {
		session.State = "initial"
		session.CandidateID = ""
		return "CON Vote cancelled. Returning to main menu."
	}

	return "CON Invalid option. Please select 1 to confirm or 2 to cancel."
}

func (s *Service) handleCompletedState(session *USSDSession, text string, step int) string {
	return "END Thank you for using the Pension Voting System. Goodbye!"
}
