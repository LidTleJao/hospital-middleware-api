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
	"github.com/LidTleJao/hospital-middleware-api/internal/repository"
	"github.com/LidTleJao/hospital-middleware-api/internal/service"
)

type stubPatientService struct {
	patients []model.Patient
	patient  model.Patient
	err      error
}

func (s stubPatientService) Search(_ context.Context, _ int64, _ repository.PatientFilter, _, _ int) ([]model.Patient, error) {
	return s.patients, s.err
}

func (s stubPatientService) Import(_ context.Context, _ int64, _ string) (model.Patient, error) {
	return s.patient, s.err
}

func ptr(s string) *string { return &s }

func TestPatientImport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       string
		service    stubPatientService
		wantStatus int
		wantField  string
		wantCount  int
	}{
		{
			name:       "imports the patient",
			body:       `{"id":"P001"}`,
			service:    stubPatientService{patient: model.Patient{ID: 7, FirstNameTH: ptr("สมชาย")}},
			wantStatus: http.StatusOK,
			wantField:  "first_name_th",
		},
		{
			name:       "rejects an unknown patient",
			body:       `{"id":"P999"}`,
			service:    stubPatientService{err: errors.New("patient not found")},
			wantStatus: http.StatusNotFound,
			wantField:  "error",
		},
		{
			name:       "rejects a request without an id",
			body:       `{}`,
			service:    stubPatientService{err: errors.New("id is required")},
			wantStatus: http.StatusBadRequest,
			wantField:  "error",
		}, {
			name:       "returns an empty array when nothing matches",
			body:       `{"first_name":"ไม่มีใครชื่อนี้"}`,
			service:    stubPatientService{patients: []model.Patient{}},
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewPatient(tt.service)

			router := gin.New()
			router.POST("/patient/import", func(c *gin.Context) {
				c.Set("claims", service.Claims{HospitalID: 1})
				handler.Import(c)
			})

			req := httptest.NewRequest(http.MethodPost, "/patient/import", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code == http.StatusOK {
				var got map[string]any
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatalf("response is not JSON: %v", err)
				}
				if _, ok := got["patient_hn"]; !ok {
					t.Errorf("response has no patient_hn: %s", rec.Body.String())
				}

			} else {
				var got map[string]any
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatalf("response is not JSON: %v", err)
				}
				if _, ok := got["error"]; !ok {
					t.Errorf("error response has no error field: %s", rec.Body.String())
				}
			}

		})
	}
}

func TestPatientSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       string
		service    stubPatientService
		wantStatus int
		wantCount  int
	}{
		{
			name:       "returns a list of patients",
			body:       `{"first_name":"สมชาย"}`,
			service:    stubPatientService{patients: []model.Patient{{ID: 1}, {ID: 2}}},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name:       "rejects a malformed body",
			body:       `{"first_name": 123}`,
			service:    stubPatientService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "reports an unexpected failure as 500",
			body:       `{"first_name":"สมชาย"}`,
			service:    stubPatientService{err: errors.New("boom")},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewPatient(tt.service)

			router := gin.New()
			router.POST("/patient/search", func(c *gin.Context) {
				c.Set("claims", service.Claims{HospitalID: 1})
			}, handler.Search)

			req := httptest.NewRequest(http.MethodPost, "/patient/search", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code == http.StatusOK {
				var got []map[string]any
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatalf("response is not a JSON array: %v", err)
				}
				if len(got) != tt.wantCount {
					t.Errorf("got %d patients, want %d", len(got), tt.wantCount)
				}
			} else {
				var got map[string]any
				json.Unmarshal(rec.Body.Bytes(), &got)
				if _, ok := got["error"]; !ok {
					t.Errorf("error response has no error field: %s", rec.Body.String())
				}
			}

		})
	}
}
