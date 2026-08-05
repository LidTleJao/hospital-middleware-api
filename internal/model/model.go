// Package model holds the structs that mirror the database tables.
package model

import "time"

// Hospital is data seed into the hospitals table at boot. It is used to validate hospital codes in requests.
type Hospital struct {
	ID        int64     `db:"hospital_id" json:"id"`
	Code      string    `db:"hospital_code" json:"hospital_code"`
	NameTH    *string   `db:"hospital_name_th" json:"hospital_name_th"`
	NameEN    *string   `db:"hospital_name_en" json:"hospital_name_en"`
	CreatedAt time.Time `db:"created_at"  json:"created_at"`
	UpdatedAt time.Time `db:"updated_at"  json:"updated_at"`
}

// Patient is a patient record cached from a hospital's HIS.
type Patient struct {
	ID           int64     `db:"patient_id"  json:"id"`
	HospitalID   int64     `db:"hospital_id" json:"hospital_id"`
	FirstNameTH  *string   `db:"first_name_th" json:"first_name_th"`
	MiddleNameTH *string   `db:"middle_name_th" json:"middle_name_th"`
	LastNameTH   *string   `db:"last_name_th"  json:"last_name_th"`
	FirstNameEN  *string   `db:"first_name_en" json:"first_name_en"`
	MiddleNameEN *string   `db:"middle_name_en" json:"middle_name_en"`
	LastNameEN   *string   `db:"last_name_en"  json:"last_name_en"`
	DateOfBirth  time.Time `db:"date_of_birth" json:"date_of_birth"`
	PatientHN    string    `db:"patient_hn" json:"patient_hn"`
	NationalID   *string   `db:"national_id" json:"national_id"`
	PassportID   *string   `db:"passport_id" json:"passport_id"`
	PhoneNumber  *string   `db:"phone_number" json:"phone_number"`
	Email        *string   `db:"email"        json:"email"`
	Gender       string    `db:"gender"       json:"gender"`
	CreatedAt    time.Time `db:"created_at"  json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"  json:"updated_at"`
}

// Staff is data staff connected to a hospital. It is used to validate staff credentials in requests.
type Staff struct {
	ID           int64     `db:"staff_id"  json:"id"`
	HospitalID   int64     `db:"hospital_id" json:"hospital_id"`
	Username     string    `db:"username"  json:"username"`
	PasswordHash string    `db:"password"  json:"-"`
	CreatedAt    time.Time `db:"created_at"  json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"  json:"updated_at"`
}
