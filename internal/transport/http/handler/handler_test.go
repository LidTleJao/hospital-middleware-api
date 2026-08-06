package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/LidTleJao/hospital-middleware-api/internal/model"
	"github.com/LidTleJao/hospital-middleware-api/internal/service"
)

// stubStaffService stands in for the real service. Each field holds what the
// corresponding method should return, so a test says what it needs and nothing
// more.
type stubStaffService struct {
	staff model.Staff
	err   error
}

// stubStaffService implements the staffService interface.
func (s stubStaffService) Create(_ context.Context, _, _, _ string) (model.Staff, error) {
	return s.staff, s.err
}

// stubStaffService implements the staffService interface.
func (s stubStaffService) Login(_ context.Context, _, _, _ string) (model.Staff, error) {
	return s.staff, s.err
}

// stubTokenIssuer returns a fixed token so a test never depends on real signing.
type stubTokenIssuer struct {
	token string
	err   error
}

func (s stubTokenIssuer) Issue(_, _ int64, _ string) (string, error) {
	return s.token, s.err
}

// TestStaffCreate tests the POST /staff/create endpoint.
func TestStaffCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       string
		service    stubStaffService
		wantStatus int
		wantField  string
	}{
		{
			name:       "creates the staff",
			body:       `{"username":"alice","password":"secret123","hospital":"H001"}`,
			service:    stubStaffService{staff: model.Staff{ID: 7, Username: "alice"}},
			wantStatus: http.StatusCreated,
			wantField:  "username",
		},
		{
			name:       "rejects an unknown hospital",
			body:       `{"username":"alice","password":"secret123","hospital":"H999"}`,
			service:    stubStaffService{err: service.ErrHospitalNotFound},
			wantStatus: http.StatusNotFound,
			wantField:  "error",
		},
		{
			name:       "rejects a short password",
			body:       `{"username":"alice","password":"123","hospital":"H001"}`,
			service:    stubStaffService{},
			wantStatus: http.StatusBadRequest,
			wantField:  "error",
		},
		{
			name:       "rejects a missing hospital",
			body:       `{"username":"alice","password":"secret123"}`,
			service:    stubStaffService{},
			wantStatus: http.StatusBadRequest,
			wantField:  "error",
		},
		{
			name:       "reports an unexpected failure as 500",
			body:       `{"username":"alice","password":"secret123","hospital":"H001"}`,
			service:    stubStaffService{err: errors.New("boom")},
			wantStatus: http.StatusInternalServerError,
			wantField:  "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewStaff(tt.service, stubTokenIssuer{token: "fixed-token"})

			router := gin.New()
			router.POST("/staff/create", handler.Create)

			req := httptest.NewRequest(http.MethodPost, "/staff/create", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}

			var payload map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatalf("response is not JSON: %v", err)
			}
			if _, ok := payload[tt.wantField]; !ok {
				t.Errorf("response has no %q field: %s", tt.wantField, rec.Body.String())
			}
			if _, leaked := payload["password"]; leaked {
				t.Error("response leaked the password field")
			}
		})
	}
}

// TestStaffLogin tests the POST /staff/login endpoint.
func TestStaffLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       string
		service    stubStaffService
		wantStatus int
		wantField  string
	}{
		{
			name:       "logs in the staff",
			body:       `{"username":"alice","password":"secret123","hospital":"H001"}`,
			service:    stubStaffService{staff: model.Staff{ID: 7, Username: "alice"}},
			wantStatus: http.StatusOK,
			wantField:  "token",
		},
		{
			name:       "unknown hospital is indistinguishable from a wrong password",
			body:       `{"username":"alice","password":"secret123","hospital":"H999"}`,
			service:    stubStaffService{err: service.ErrInvalidCredentials},
			wantStatus: http.StatusUnauthorized,
			wantField:  "error",
		},
		{
			name:       "rejects invalid credentials",
			body:       `{"username":"alice","password":"wrongpass","hospital":"H001"}`,
			service:    stubStaffService{err: service.ErrInvalidCredentials},
			wantStatus: http.StatusUnauthorized,
			wantField:  "error",
		},
		{
			name:       "reports an unexpected failure as 500",
			body:       `{"username":"alice","password":"secret123","hospital":"H001"}`,
			service:    stubStaffService{err: errors.New("boom")},
			wantStatus: http.StatusInternalServerError,
			wantField:  "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewStaff(tt.service, stubTokenIssuer{token: "fixed-token"})

			router := gin.New()
			router.POST("/staff/login", handler.Login)

			req := httptest.NewRequest(http.MethodPost, "/staff/login", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}

			var payload map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatalf("response is not JSON: %v", err)
			}
			if _, ok := payload[tt.wantField]; !ok {
				t.Errorf("response has no %q field: %s", tt.wantField, rec.Body.String())
			}
			if _, leaked := payload["password"]; leaked {
				t.Error("response leaked the password field")
			}
		})
	}
}
