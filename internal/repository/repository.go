// Package repository turns SQL rows into model values and back.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/LidTleJao/hospital-middleware-api/internal/model"
)

// ErrNotFound is returned when a lookup matches no row.
var ErrNotFound = errors.New("not found")

// Hospital reads and writes rows in the hospitals table.
type Hospital struct {
	db *sqlx.DB
}

// NewHospital returns a Hospital backed by db.
func NewHospital(db *sqlx.DB) *Hospital {
	return &Hospital{db: db}
}

// FindByCode returns the hospital with the given code, or ErrNotFound if no such hospital exists.
func (r *Hospital) FindByCode(ctx context.Context, code string) (model.Hospital, error) {
	const query = `
		SELECT hospital_id, hospital_code, hospital_name_th, hospital_name_en, created_at, updated_at
		FROM hospitals
		WHERE hospital_code = $1`

	var hospital model.Hospital
	if err := r.db.GetContext(ctx, &hospital, query, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Hospital{}, ErrNotFound
		}
		return model.Hospital{}, fmt.Errorf("find hospital by code: %w", err)
	}
	return hospital, nil
}

// Staff reads and writes rows in the staff table.
type Staff struct {
	db *sqlx.DB
}

// NewStaff returns a Staffs backed by db.
func NewStaff(db *sqlx.DB) *Staff {
	return &Staff{db: db}
}

// Create inserts a staffs and returns it with the generated id filled in.
func (r *Staff) Create(ctx context.Context, hospitalID int64, username, passwordHash string) (model.Staff, error) {
	const query = `
		INSERT INTO staffs (hospital_id, username, password)
		VALUES ($1, $2, $3)
		RETURNING staff_id, hospital_id, username, password, created_at, updated_at`

	var staff model.Staff
	if err := r.db.GetContext(ctx, &staff, query, hospitalID, username, passwordHash); err != nil {
		return model.Staff{}, fmt.Errorf("create staff: %w", err)
	}
	return staff, nil
}

// FindByUsername returns the staff owning username inside hospitalID.
func (r *Staff) FindByUsername(ctx context.Context, hospitalID int64, username string) (model.Staff, error) {
	const query = `
		SELECT staff_id, hospital_id, username, password, created_at, updated_at
		FROM staffs
		WHERE hospital_id = $1 AND username = $2`

	var staff model.Staff
	if err := r.db.GetContext(ctx, &staff, query, hospitalID, username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Staff{}, ErrNotFound
		}
		return model.Staff{}, fmt.Errorf("find staff by username: %w", err)
	}
	return staff, nil
}
