// Package service holds the business rules that sit between HTTP and SQL.
package service

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/LidTleJao/hospital-middleware-api/internal/model"
	"github.com/LidTleJao/hospital-middleware-api/internal/repository"
)

// Errors the HTTP layer translates into status codes.
var (
	ErrHospitalNotFound   = errors.New("hospital not found")
	ErrInvalidCredentials = errors.New("invalid username or password")
)

// hospitalFinder is the slice of the hospital repository this service uses.
// It is declared here, in the consumer, so the service can be tested with a
// stub and never has to import a database driver.
type hospitalFinder interface {
	FindByCode(ctx context.Context, code string) (model.Hospital, error)
}

// staffStore is the slice of the staff repository this service uses.
type staffStore interface {
	Create(ctx context.Context, hospitalID int64, username, passwordHash string) (model.Staff, error)
	FindByUsername(ctx context.Context, hospitalID int64, username string) (model.Staff, error)
}

// Staff is the service that creates and authenticates staff.
type Staff struct {
	hospitals hospitalFinder
	staffs    staffStore
}

// NewStaff wires the service to its dependencies.
func NewStaff(hospital hospitalFinder, staff staffStore) *Staff {
	return &Staff{hospitals: hospital, staffs: staff}
}

// Create creates a staff under the hospital identified by code.
func (s *Staff) Create(ctx context.Context, code, username, password string) (model.Staff, error) {
	hospital, err := s.hospitals.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Staff{}, ErrHospitalNotFound
		}
		return model.Staff{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.Staff{}, fmt.Errorf("hash password: %w", err)
	}

	return s.staffs.Create(ctx, hospital.ID, username, string(hash))
}

// Login checks a login. Every failure returns the same error so the
// caller cannot tell a wrong password from a username that does not exist.
func (s *Staff) Login(ctx context.Context, code, username, password string) (model.Staff, error) {
	hospital, err := s.hospitals.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Staff{}, ErrInvalidCredentials
		}
		return model.Staff{}, err
	}

	staff, err := s.staffs.FindByUsername(ctx, hospital.ID, username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Staff{}, ErrInvalidCredentials
		}
		return model.Staff{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(staff.PasswordHash), []byte(password)); err != nil {
		return model.Staff{}, ErrInvalidCredentials
	}

	return staff, nil
}
