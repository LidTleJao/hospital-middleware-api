// Package hisclient talks to the external HIS service.
package hisclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ErrNotFound is returned when the HIS holds no record for the id.
var ErrNotFound = errors.New("not found")

// Patient is the HIS record for a patient. It is a subset of the HIS's
// full record, and it is used to avoid leaking sensitive information like the
type Patient struct {
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

// Client calls the HIS service.
type Client struct {
	baseURL string
	http    *http.Client
}

// New returns a Client that calls the HIS service at baseURL with timeout.
func New(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
	}
}

// FetchByID calls the HIS service to get the patient record for id.
func (c *Client) FetchByID(ctx context.Context, id string) (Patient, error) {
	endpoint := c.baseURL + "/patient/search/" + url.PathEscape(id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Patient{}, fmt.Errorf("build HIS request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return Patient{}, fmt.Errorf("call HIS: %w", err)
	}
	// Always close the body, or the connection leaks out of the pool.
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return Patient{}, ErrNotFound
	default:
		return Patient{}, fmt.Errorf("HIS returned %s", resp.Status)
	}

	var patient Patient
	if err := json.NewDecoder(resp.Body).Decode(&patient); err != nil {
		return Patient{}, fmt.Errorf("decode HIS response: %w", err)
	}

	return patient, nil
}
