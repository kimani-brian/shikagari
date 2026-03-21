package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shikagari/api/internal/dto"
	"github.com/shikagari/api/internal/middleware"
)

func TestAuthHandler_ListSecurityEvents_ReturnsEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	events := []dto.SecurityEventResponse{
		{ID: "evt-1", Label: "Login", Timestamp: "2024-03-10T10:00:00Z", Severity: "info", Details: "MacBook"},
	}

	svc := &mockAuthService{securityEvents: events}
	handler := NewAuthHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/auth/security-events", nil)
	c.Set(middleware.ContextUserKey, &middleware.Claims{UserID: uuid.New(), SessionID: uuid.New()})

	handler.ListSecurityEvents(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var payload struct {
		Success bool                        `json:"success"`
		Data    []dto.SecurityEventResponse `json:"data"`
		Message string                      `json:"message"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !payload.Success {
		t.Fatalf("expected success to be true")
	}

	if len(payload.Data) != len(events) {
		t.Fatalf("expected %d events, got %d", len(events), len(payload.Data))
	}

	if payload.Data[0].Label != events[0].Label {
		t.Fatalf("expected event label %q, got %q", events[0].Label, payload.Data[0].Label)
	}
}

func TestAuthHandler_ListSessions_ReturnsEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sessionID := uuid.New()
	sessions := []dto.SessionResponse{
		{ID: sessionID.String(), Device: "Mac", Browser: "Chrome", Location: "Nairobi", LastActive: "2024-03-10T11:00:00Z", Current: true},
	}

	svc := &mockAuthService{sessions: sessions}
	handler := NewAuthHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/auth/sessions", nil)
	c.Set(middleware.ContextUserKey, &middleware.Claims{UserID: uuid.New(), SessionID: sessionID})

	handler.ListSessions(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var payload struct {
		Success bool                  `json:"success"`
		Data    []dto.SessionResponse `json:"data"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(payload.Data) != len(sessions) {
		t.Fatalf("expected %d sessions, got %d", len(sessions), len(payload.Data))
	}

	if payload.Data[0].ID != sessions[0].ID {
		t.Fatalf("expected session id %q, got %q", sessions[0].ID, payload.Data[0].ID)
	}
}

type mockAuthService struct {
	securityEvents []dto.SecurityEventResponse
	sessions       []dto.SessionResponse
}

func (m *mockAuthService) Register(dto.RegisterRequest, dto.SessionMetadata) (*dto.AuthResponse, error) {
	return nil, nil
}

func (m *mockAuthService) Login(dto.LoginRequest, dto.SessionMetadata) (*dto.AuthResponse, error) {
	return nil, nil
}

func (m *mockAuthService) RequestPasswordReset(dto.PasswordResetRequest) error {
	return nil
}

func (m *mockAuthService) ResetPassword(dto.PasswordResetConfirmRequest) error {
	return nil
}

func (m *mockAuthService) ListSecurityEvents(uuid.UUID) ([]dto.SecurityEventResponse, error) {
	return m.securityEvents, nil
}

func (m *mockAuthService) ListSessions(uuid.UUID, uuid.UUID) ([]dto.SessionResponse, error) {
	return m.sessions, nil
}

func (m *mockAuthService) RevokeSession(uuid.UUID, uuid.UUID) error {
	return nil
}

func (m *mockAuthService) Logout(uuid.UUID, uuid.UUID) error {
	return nil
}
