package service_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"agnos-assessment/internal/model"
	"agnos-assessment/internal/service"

	"github.com/stretchr/testify/assert"
)

func TestHISClient_FetchPatientFromHospitalA_Success(t *testing.T) {
	// Create mock HTTP server simulating Hospital A HIS API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/patient/search/1100501234567", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		patient := model.Patient{
			PatientHN:   "HN-A-9999",
			NationalID:  "1100501234567",
			FirstNameTH: "ประสิทธิ์",
			LastNameTH:  "โชคดี",
			FirstNameEN: "Prasit",
			LastNameEN:  "Chokdee",
			DateOfBirth: "1985-06-20",
			Gender:      "M",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(patient)
	}))
	defer server.Close()

	client := service.NewHISClient(server.URL)
	patient, err := client.FetchPatientFromHospitalA("1100501234567")

	assert.NoError(t, err)
	assert.NotNil(t, patient)
	assert.Equal(t, "HN-A-9999", patient.PatientHN)
	assert.Equal(t, "hospital-a", patient.Hospital)
	assert.Equal(t, "Prasit", patient.FirstNameEN)
}

func TestHISClient_FetchPatientFromHospitalA_NotFound_404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := service.NewHISClient(server.URL)
	patient, err := client.FetchPatientFromHospitalA("0000000000000")

	assert.NoError(t, err)
	assert.Nil(t, patient)
}

func TestHISClient_FetchPatientFromHospitalA_ServerError_500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := service.NewHISClient(server.URL)
	patient, err := client.FetchPatientFromHospitalA("1100501234567")

	assert.Error(t, err)
	assert.Nil(t, patient)
	assert.Contains(t, err.Error(), "unexpected status: 500")
}

func TestHISClient_FetchPatientFromHospitalA_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{invalid-json"))
	}))
	defer server.Close()

	client := service.NewHISClient(server.URL)
	patient, err := client.FetchPatientFromHospitalA("1100501234567")

	assert.Error(t, err)
	assert.Nil(t, patient)
	assert.Contains(t, err.Error(), "failed to decode HIS response")
}

func TestHISClient_FetchPatientFromHospitalA_NetworkError(t *testing.T) {
	// Point to an invalid closed port to trigger network error
	client := service.NewHISClient("http://127.0.0.1:54321")
	patient, err := client.FetchPatientFromHospitalA("1100501234567")

	assert.Error(t, err)
	assert.Nil(t, patient)
	assert.Contains(t, err.Error(), "HIS API request failed")
}

func TestHISClient_FetchPatientFromHospitalA_InvalidURL(t *testing.T) {
	// A URL with control character causes http.NewRequest to fail immediately
	client := service.NewHISClient("http://localhost\n")
	patient, err := client.FetchPatientFromHospitalA("1100501234567")

	assert.Error(t, err)
	assert.Nil(t, patient)
	assert.Contains(t, err.Error(), "failed to create HIS request")
}


