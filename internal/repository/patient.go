package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/LidTleJao/hospital-middleware-api/internal/model"
)

// Patient reads and writes rows in Patient table.
type Patient struct {
	db *sqlx.DB
}

// NewPatient returns a Patient backed by db.
func NewPatient(db *sqlx.DB) *Patient {
	return &Patient{db: db}
}

// PatientFilter is the set of optional criteria a search may narrow by. A nil
// field means "the caller did not ask about this", which is different from
// asking for an empty value.
type PatientFilter struct {
	FirstName   *string
	MiddleName  *string
	LastName    *string
	DateOfBirth *string
	NationalID  *string
	PassportID  *string
	PhoneNumber *string
	Email       *string
}

// Upsert stores a patient, replacing the row that already holds the same
// (publisher_id, patient_hn) rather than failing on the unique constraint.
func (r *Patient) Upsert(ctx context.Context, patient model.Patient) (model.Patient, error) {
	const query = `
	INSERT INTO patients (hospital_id, first_name_th, first_name_en, middle_name_th,  middle_name_en, last_name_th, last_name_en, date_of_birth, patient_hn, national_id, passport_id, phone_number, email, gender)
	VALUES (:hospital_id, :first_name_th, :first_name_en, :middle_name_th, :middle_name_en, :last_name_th, :last_name_en, :date_of_birth, :patient_hn, :national_id, :passport_id, :phone_number, :email, :gender)
	ON CONFLICT (hospital_id, patient_hn) DO UPDATE
	SET first_name_th = EXCLUDED.first_name_th,
		first_name_en = EXCLUDED.first_name_en,
		middle_name_th = EXCLUDED.middle_name_th,
		middle_name_en = EXCLUDED.middle_name_en,
		last_name_th = EXCLUDED.last_name_th,
		last_name_en = EXCLUDED.last_name_en,
		date_of_birth = EXCLUDED.date_of_birth,
		national_id = EXCLUDED.national_id,
		passport_id = EXCLUDED.passport_id,
		phone_number = EXCLUDED.phone_number,
		email = EXCLUDED.email,
		gender = EXCLUDED.gender
	RETURNING patient_id, hospital_id, first_name_th, first_name_en, middle_name_th, middle_name_en, last_name_th, last_name_en, date_of_birth, patient_hn, national_id, passport_id, phone_number, email, gender, created_at, updated_at`

	rows, err := r.db.NamedQueryContext(ctx, query, patient)
	if err != nil {
		return model.Patient{}, fmt.Errorf("upsert patient: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return model.Patient{}, fmt.Errorf("upsert patient: no row returned")
	}

	var stored model.Patient
	if err := rows.StructScan(&stored); err != nil {
		return model.Patient{}, fmt.Errorf("scan upserted patient: %w", err)
	}
	return stored, nil
}

// Search returns the patients of one hospital that match every criterion the
// caller supplied. hospitalID is never taken from the request, so a caller
// cannot reach another hospital's rows.
func (r *Patient) Search(ctx context.Context, hospitalID int64, filter PatientFilter, limit, offset int) ([]model.Patient, error) {
	where := []string{"hospital_id = $1"}
	args := []any{hospitalID}

	// exact adds "column = $n" and remembers the value behind it.
	exact := func(column string, value *string) {
		if value == nil {
			return
		}
		args = append(args, *value)
		where = append(where, fmt.Sprintf("%s = $%d", column, len(args)))
	}

	// either matches value against two columns with a single placeholder. The
	// outer parentheses are required: without them the OR would escape this
	// condition and combine with "publisher_id = $1", widening the search
	// beyond the caller's own publisher.
	either := func(columnA, columnB string, value *string) {
		if value == nil {
			return
		}
		args = append(args, *value)
		where = append(where, fmt.Sprintf(
			"(%s ILIKE '%%' || $%d || '%%' OR %s ILIKE '%%' || $%d || '%%')",
			columnA, len(args), columnB, len(args)))
	}

	either("first_name_th", "first_name_en", filter.FirstName)
	either("middle_name_th", "middle_name_en", filter.MiddleName)
	either("last_name_th", "last_name_en", filter.LastName)
	exact("date_of_birth", filter.DateOfBirth)
	exact("national_id", filter.NationalID)
	exact("passport_id", filter.PassportID)
	exact("phone_number", filter.PhoneNumber)
	exact("email", filter.Email)

	args = append(args, limit, offset)

	query := fmt.Sprintf(`
	SELECT patient_id, hospital_id, first_name_th, first_name_en, middle_name_th, middle_name_en, last_name_th, last_name_en, date_of_birth, patient_hn, national_id, passport_id, phone_number, email, gender, created_at, updated_at
	FROM patients
	WHERE %s
	ORDER BY patient_id
	LIMIT $%d OFFSET $%d`, strings.Join(where, " AND "), len(args)-1, len(args))

	patients := []model.Patient{}
	if err := r.db.SelectContext(ctx, &patients, query, args...); err != nil {
		return nil, fmt.Errorf("search patients: %w", err)
	}
	return patients, nil
}
