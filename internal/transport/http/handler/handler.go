// Package handler turns HTTP requests into service calls and back.
package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/LidTleJao/hospital-middleware-api/internal/model"
	"github.com/LidTleJao/hospital-middleware-api/internal/service"
)

// staffService is the slice of the staff service this handler uses.
type staffService interface {
	Create(ctx context.Context, code, username, password string) (model.Staff, error)
	Login(ctx context.Context, code, username, password string) (model.Staff, error)
}

// Staff serves the staff endpoints.
type Staff struct {
	staffs staffService
}

// NewStaff returns a handler backed by svc.
func NewStaff(svc staffService) *Staff {
	return &Staff{staffs: svc}
}

// createRequest is the body POST /staff/create accepts.
type createRequest struct {
	Username     string `json:"username"  binding:"required,min=3,max=64"`
	Password     string `json:"password"  binding:"required,min=8"`
	HospitalCode string `json:"hospital" binding:"required,max=50"`
}

// staffResponse is the shape of the staff returned to the client. It is a subset of
// model.Staff, and it is used to avoid leaking sensitive information like the password hash.
type staffResponse struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	HospitalCode string `json:"hospital"`
}

// errorResponse is the single error shape every endpoint returns.
type errorResponse struct {
	Error string `json:"error"`
}

// Create handles POST /staff/create.
func (h *Staff) Create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	staff, err := h.staffs.Create(c.Request.Context(), req.HospitalCode, req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrHospitalNotFound):
			c.JSON(http.StatusNotFound, errorResponse{Error: "hospital not found"})
		default:
			c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal error"})
		}
		return
	}

	c.JSON(http.StatusCreated, staffResponse{
		ID:           staff.ID,
		Username:     staff.Username,
		HospitalCode: req.HospitalCode,
	})
}

// loginRequest is the body POST /staff/login accepts.
type loginRequest struct {
	Username     string `json:"username"  binding:"required,min=3,max=64"`
	Password     string `json:"password"  binding:"required,min=8"`
	HospitalCode string `json:"hospital" binding:"required,max=50"`
}

// Login handles POST /staff/login.
func (h *Staff) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	staff, err := h.staffs.Login(c.Request.Context(), req.HospitalCode, req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, errorResponse{Error: "invalid username or password"})
		default:
			c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal error"})
		}
		return
	}

	c.JSON(http.StatusOK, staffResponse{
		ID:           staff.ID,
		Username:     staff.Username,
		HospitalCode: req.HospitalCode,
	})
}
