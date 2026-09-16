package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"agnos-assessment/internal/model"
)

// HISClient defines the interface for communicating with external Hospital Information Systems
type HISClient interface {
	FetchPatientFromHospitalA(id string) (*model.Patient, error)
}

type hisClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHISClient(baseURL string) HISClient {
	return &hisClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// FetchPatientFromHospitalA calls GET https://hospital-a.api.co.th/patient/search/{id}
func (c *hisClient) FetchPatientFromHospitalA(id string) (*model.Patient, error) {
	url := fmt.Sprintf("%s/patient/search/%s", c.baseURL, id)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HIS request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HIS API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HIS API returned unexpected status: %d", resp.StatusCode)
	}

	var patient model.Patient
	if err := json.NewDecoder(resp.Body).Decode(&patient); err != nil {
		return nil, fmt.Errorf("failed to decode HIS response: %w", err)
	}

	// Assign hospital scope for Hospital A
	patient.Hospital = "hospital-a"
	return &patient, nil
}

// MockHISClient for unit tests and local development fallback
type MockHISClient struct {
	MockData map[string]*model.Patient
}

func NewMockHISClient() *MockHISClient {
	return &MockHISClient{
		MockData: make(map[string]*model.Patient),
	}
}

func (m *MockHISClient) FetchPatientFromHospitalA(id string) (*model.Patient, error) {
	if patient, ok := m.MockData[id]; ok {
		// Return copy
		p := *patient
		p.Hospital = "hospital-a"
		return &p, nil
	}
	return nil, nil
}
