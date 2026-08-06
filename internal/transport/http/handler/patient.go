// Package handler turns HTTP requests into service calls and back.
package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/LidTleJao/hospital-middleware-api/internal/model"
	"github.com/LidTleJao/hospital-middleware-api/internal/repository"
	"github.com/LidTleJao/hospital-middleware-api/internal/service"
	"github.com/LidTleJao/hospital-middleware-api/internal/transport/http/middleware"
)

// patientService is the slice of the patient service this handler uses.
type patientService interface {
	Import(ctx context.Context, hospitalID int64, id string) (model.Patient, error)
	Search(ctx context.Context, hospitalID int64, filter repository.PatientFilter, pageSize, page int) ([]model.Patient, error)
}

// Patient serves the patient endpoints.
type Patient struct {
	patients patientService
}

// syncRequest is the body POST /patients/sync accepts.
type syncRequest struct {
	ID string `json:"id" binding:"required"`
}

// NewPatient returns a handler backed by svc.
func NewPatient(svc patientService) *Patient {
	return &Patient{patients: svc}
}

// searchRequest is the body GET /patients/search accepts.
type searchRequest struct {
	NationalID  *string `json:"national_id"`
	PassportID  *string `json:"passport_id"`
	FirstName   *string `json:"first_name"`
	MiddleName  *string `json:"middle_name"`
	LastName    *string `json:"last_name"`
	DateOfBirth *string `json:"date_of_birth"`
	PhoneNumber *string `json:"phone_number"`
	Email       *string `json:"email"`
	Page        int     `json:"page"`
	PageSize    int     `json:"page_size"`
}

// patientResponse is the shape of the patient returned to the client. It is a subset of
// model.Patient, and it is used to avoid leaking sensitive information like the patient HN.
type patientResponse struct {
	ID           int64   `json:"id"`
	HospitalID   int64   `json:"hospital_id"`
	FirstNameTH  *string `json:"first_name_th"`
	MiddleNameTH *string `json:"middle_name_th"`
	LastNameTH   *string `json:"last_name_th"`
	FirstNameEN  *string `json:"first_name_en"`
	MiddleNameEN *string `json:"middle_name_en"`
	LastNameEN   *string `json:"last_name_en"`
	DateOfBirth  string  `json:"date_of_birth"`
	PatientHN    string  `json:"patient_hn"`
	NationalID   *string `json:"national_id"`
	PassportID   *string `json:"passport_id"`
	PhoneNumber  *string `json:"phone_number"`
	Email        *string `json:"email"`
	Gender       string  `json:"gender"`
}

// Search handles GET /patients/search. It returns a list of patients matching the search criteria.
func (h *Patient) Search(c *gin.Context) {
	var req searchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	claims, ok := middleware.ClaimsFrom(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	patients, err := h.patients.Search(c.Request.Context(), claims.HospitalID, repository.PatientFilter{
		FirstName:   req.FirstName,
		MiddleName:  req.MiddleName,
		LastName:    req.LastName,
		DateOfBirth: req.DateOfBirth,
		NationalID:  req.NationalID,
		PassportID:  req.PassportID,
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
	}, req.PageSize, req.Page)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrPatientNotFound):
			c.JSON(http.StatusNotFound, errorResponse{Error: "patient not found"})
		default:
			c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal error"})
		}
		return
	}

	res := make([]patientResponse, 0, len(patients))
	for _, p := range patients {
		res = append(res, patientResponse{
			ID:           p.ID,
			HospitalID:   p.HospitalID,
			FirstNameTH:  p.FirstNameTH,
			MiddleNameTH: p.MiddleNameTH,
			LastNameTH:   p.LastNameTH,
			FirstNameEN:  p.FirstNameEN,
			MiddleNameEN: p.MiddleNameEN,
			LastNameEN:   p.LastNameEN,
			DateOfBirth:  p.DateOfBirth.Format("2006-01-02"),
			PatientHN:    p.PatientHN,
			NationalID:   p.NationalID,
			PassportID:   p.PassportID,
			PhoneNumber:  p.PhoneNumber,
			Email:        p.Email,
			Gender:       p.Gender,
		})
	}

	c.JSON(http.StatusOK, res)
}

func (h *Patient) Import(c *gin.Context) {
	var req syncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	claims, ok := middleware.ClaimsFrom(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	patient, err := h.patients.Import(c.Request.Context(), claims.HospitalID, req.ID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPatientNotFound):
			c.JSON(http.StatusNotFound, errorResponse{Error: "patient not found"})
		default:
			c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal error"})
		}
		return
	}

	res := patientResponse{
		ID:           patient.ID,
		HospitalID:   patient.HospitalID,
		FirstNameTH:  patient.FirstNameTH,
		MiddleNameTH: patient.MiddleNameTH,
		LastNameTH:   patient.LastNameTH,
		FirstNameEN:  patient.FirstNameEN,
		MiddleNameEN: patient.MiddleNameEN,
		LastNameEN:   patient.LastNameEN,
		DateOfBirth:  patient.DateOfBirth.Format("2006-01-02"),
		PatientHN:    patient.PatientHN,
		NationalID:   patient.NationalID,
		PassportID:   patient.PassportID,
		PhoneNumber:  patient.PhoneNumber,
		Email:        patient.Email,
		Gender:       patient.Gender,
	}

	c.JSON(http.StatusOK, res)
}
