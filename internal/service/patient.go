package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/LidTleJao/hospital-middleware-api/internal/hisclient"
	"github.com/LidTleJao/hospital-middleware-api/internal/model"
	"github.com/LidTleJao/hospital-middleware-api/internal/repository"
)

// ErrPatientNotFound is returned when a patient record is not found in the HIS or local repository.
var ErrPatientNotFound = errors.New("patient not found")

// Pagination bounds applied to every search.
const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// hisFetcher is the interface the service uses to fetch hospital records from the HIS.
type hisFetcher interface {
	FetchByID(ctx context.Context, id string) (hisclient.Patient, error)
}

// patientStore is the slice of the patient repository this service uses.
type patientStore interface {
	Upsert(ctx context.Context, patient model.Patient) (model.Patient, error)
	Search(ctx context.Context, hospitalID int64, filter repository.PatientFilter, limit, offset int) ([]model.Patient, error)
}

// Patient is the service that handles patient records. It fetches records from the HIS and upserts them into the local repository.
type Patient struct {
	his      hisFetcher
	patients patientStore
}

// NewPatient wires the service to its dependencies.
func NewPatient(his hisFetcher, patients patientStore) *Patient {
	return &Patient{
		his:      his,
		patients: patients,
	}
}

// Import fetches a patient record from the HIS by ID and upserts it into the local repository. It returns the upserted record or an error.
func (s *Patient) Import(ctx context.Context, hospitalID int64, id string) (model.Patient, error) {
	fetched, err := s.his.FetchByID(ctx, id)
	if err != nil {
		if errors.Is(err, hisclient.ErrNotFound) {
			return model.Patient{}, ErrPatientNotFound
		}
		return model.Patient{}, err
	}

	dateOfBirth, err := time.Parse(time.DateOnly, fetched.DateOfBirth)
	if err != nil {
		return model.Patient{}, fmt.Errorf("parse date of birth: %w", err)
	}

	return s.patients.Upsert(ctx, model.Patient{
		HospitalID:   hospitalID,
		FirstNameTH:  fetched.FirstNameTH,
		MiddleNameTH: fetched.MiddleNameTH,
		LastNameTH:   fetched.LastNameTH,
		FirstNameEN:  fetched.FirstNameEN,
		MiddleNameEN: fetched.MiddleNameEN,
		LastNameEN:   fetched.LastNameEN,
		DateOfBirth:  dateOfBirth,
		PatientHN:    fetched.PatientHN,
		NationalID:   fetched.NationalID,
		PassportID:   fetched.PassportID,
		Gender:       fetched.Gender,
		Email:        fetched.Email,
		PhoneNumber:  fetched.PhoneNumber,
	})
}

// Search looks up patients in the local repository by hospital ID and filter criteria. It applies pagination bounds to the search.
func (s *Patient) Search(ctx context.Context, hospitalID int64, filter repository.PatientFilter, pageSize, page int) ([]model.Patient, error) {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	if page < 1 {
		page = 1
	}

	return s.patients.Search(ctx, hospitalID, filter, pageSize, (page-1)*pageSize)
}
