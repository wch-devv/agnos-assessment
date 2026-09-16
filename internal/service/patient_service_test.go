package service_test

import (
	"errors"
	"strings"
	"testing"

	"agnos-assessment/internal/model"
	"agnos-assessment/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockPatientRepository struct {
	patients  []model.Patient
	searchErr error
	upsertErr error
}

func newMockPatientRepository() *mockPatientRepository {
	return &mockPatientRepository{
		patients: make([]model.Patient, 0),
	}
}

func (m *mockPatientRepository) Create(p *model.Patient) error {
	m.patients = append(m.patients, *p)
	return nil
}

func (m *mockPatientRepository) Upsert(p *model.Patient) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	for i, existing := range m.patients {
		if existing.Hospital == p.Hospital && existing.PatientHN == p.PatientHN {
			m.patients[i] = *p
			return nil
		}
	}
	m.patients = append(m.patients, *p)
	return nil
}

func (m *mockPatientRepository) FindByHNAndHospital(hn, hospital string) (*model.Patient, error) {
	for _, p := range m.patients {
		if p.Hospital == hospital && p.PatientHN == hn {
			return &p, nil
		}
	}
	return nil, nil
}

func (m *mockPatientRepository) Search(hospital string, params *model.PatientSearchParams) ([]model.Patient, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	var results []model.Patient
	for _, p := range m.patients {
		// Strict hospital boundary enforcement
		if p.Hospital != hospital {
			continue
		}

		if params != nil {
			if params.NationalID != "" && p.NationalID != params.NationalID {
				continue
			}
			if params.PassportID != "" && p.PassportID != params.PassportID {
				continue
			}
			if params.FirstName != "" && !strings.Contains(strings.ToLower(p.FirstNameTH+p.FirstNameEN), strings.ToLower(params.FirstName)) {
				continue
			}
			if params.LastName != "" && !strings.Contains(strings.ToLower(p.LastNameTH+p.LastNameEN), strings.ToLower(params.LastName)) {
				continue
			}
			if params.PhoneNumber != "" && p.PhoneNumber != params.PhoneNumber {
				continue
			}
			if params.Email != "" && !strings.Contains(strings.ToLower(p.Email), strings.ToLower(params.Email)) {
				continue
			}
		}

		results = append(results, p)
	}
	return results, nil
}

func setupPatientService() (service.PatientService, *mockPatientRepository, *service.MockHISClient) {
	repo := newMockPatientRepository()
	mockHIS := service.NewMockHISClient()

	// Seed patient in hospital-a
	repo.Create(&model.Patient{
		ID:          uuid.New(),
		Hospital:    "hospital-a",
		PatientHN:   "HN-A-001",
		NationalID:  "1100501234567",
		PassportID:  "AA1234567",
		FirstNameTH: "สมชาย",
		LastNameTH:  "ใจดี",
		FirstNameEN: "Somchai",
		LastNameEN:  "Jaidee",
		DateOfBirth: "1990-05-15",
		PhoneNumber: "0812345678",
		Email:       "somchai@example.com",
		Gender:      "M",
	})

	// Seed patient in hospital-b
	repo.Create(&model.Patient{
		ID:          uuid.New(),
		Hospital:    "hospital-b",
		PatientHN:   "HN-B-001",
		NationalID:  "1200901234567",
		PassportID:  "BB7654321",
		FirstNameTH: "อนันต์",
		LastNameTH:  "สุขเจริญ",
		FirstNameEN: "Anan",
		LastNameEN:  "Sukcharoen",
		DateOfBirth: "1988-02-14",
		PhoneNumber: "0855551234",
		Email:       "anan@example.com",
		Gender:      "M",
	})

	svc := service.NewPatientService(repo, mockHIS)
	return svc, repo, mockHIS
}

// =========================================================================
// Positive Test Scenarios
// =========================================================================

func TestPatientService_Search_ByNationalID_Success(t *testing.T) {
	svc, _, _ := setupPatientService()

	params := &model.PatientSearchParams{
		NationalID: "1100501234567",
	}
	results, err := svc.SearchPatients("hospital-a", params)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "HN-A-001", results[0].PatientHN)
	assert.Equal(t, "hospital-a", results[0].Hospital)
}

func TestPatientService_Search_ByName_Success(t *testing.T) {
	svc, _, _ := setupPatientService()

	params := &model.PatientSearchParams{
		FirstName: "Somchai",
	}
	results, err := svc.SearchPatients("hospital-a", params)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Somchai", results[0].FirstNameEN)
}

func TestPatientService_HISIntegration_Fallback_Success(t *testing.T) {
	svc, repo, mockHIS := setupPatientService()

	// Mock patient in Hospital A external HIS
	hisPatientID := "1999900001111"
	mockHIS.MockData[hisPatientID] = &model.Patient{
		ID:          uuid.New(),
		PatientHN:   "HN-HIS-999",
		NationalID:  hisPatientID,
		FirstNameTH: "วิชัย",
		LastNameTH:  "มั่นคง",
		FirstNameEN: "Wichai",
		LastNameEN:  "Mankong",
		DateOfBirth: "1992-08-10",
		Gender:      "M",
	}

	params := &model.PatientSearchParams{
		NationalID: hisPatientID,
	}

	// Search by hospital-a staff should fetch from HIS and return
	results, err := svc.SearchPatients("hospital-a", params)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "HN-HIS-999", results[0].PatientHN)
	assert.Equal(t, "hospital-a", results[0].Hospital)

	// Verify cached to local repo
	cached, _ := repo.FindByHNAndHospital("HN-HIS-999", "hospital-a")
	assert.NotNil(t, cached)
}

// =========================================================================
// Negative & Boundary Test Scenarios (Cross-Hospital Isolation)
// =========================================================================

func TestPatientService_HospitalDataIsolation_StrictlyEnforced(t *testing.T) {
	svc, _, _ := setupPatientService()

	// Staff from hospital-b searches with NationalID of patient in hospital-a
	params := &model.PatientSearchParams{
		NationalID: "1100501234567", // This is hospital-a patient
	}

	// Searching from hospital-b must NEVER return hospital-a patient!
	results, err := svc.SearchPatients("hospital-b", params)
	assert.NoError(t, err)
	assert.Empty(t, results, "Staff from hospital-b MUST NOT see patient from hospital-a")
}

func TestPatientService_Search_NotFound(t *testing.T) {
	svc, _, _ := setupPatientService()

	params := &model.PatientSearchParams{
		NationalID: "0000000000000",
	}
	results, err := svc.SearchPatients("hospital-a", params)
	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestPatientService_Search_RepoError(t *testing.T) {
	repo := newMockPatientRepository()
	repo.searchErr = errors.New("database connection failed")
	svc := service.NewPatientService(repo, service.NewMockHISClient())

	results, err := svc.SearchPatients("hospital-a", &model.PatientSearchParams{})
	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestPatientService_Search_WithPassportID_CallsHIS_Success(t *testing.T) {
	svc, _, mockHIS := setupPatientService()

	passportID := "PASSPORT999"
	mockHIS.MockData[passportID] = &model.Patient{
		ID:          uuid.New(),
		PatientHN:   "HN-PASSPORT",
		PassportID:  passportID,
		FirstNameTH: "สมศรี",
		LastNameTH:  "ดีเลิศ",
		FirstNameEN: "Somsri",
		LastNameEN:  "Deelert",
		Gender:      "F",
	}

	params := &model.PatientSearchParams{
		PassportID: passportID,
	}

	results, err := svc.SearchPatients("hospital-a", params)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "HN-PASSPORT", results[0].PatientHN)
	assert.Equal(t, "hospital-a", results[0].Hospital)
}

func TestPatientService_Search_HISError_GracefulFallback(t *testing.T) {
	svc, _, mockHIS := setupPatientService()
	mockHIS.Err = errors.New("HIS server timeout")

	params := &model.PatientSearchParams{
		NationalID: "9999999999999",
	}
	// Even if HIS fails, system logs warning and returns empty results without crashing
	results, err := svc.SearchPatients("hospital-a", params)
	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestPatientService_Search_RepoUpsertError_GracefulFallback(t *testing.T) {
	repo := newMockPatientRepository()
	repo.upsertErr = errors.New("disk full")
	mockHIS := service.NewMockHISClient()

	mockHIS.MockData["1111111111111"] = &model.Patient{
		ID:         uuid.New(),
		PatientHN:  "HN-DISK-FULL",
		NationalID: "1111111111111",
		Gender:     "M",
	}

	svc := service.NewPatientService(repo, mockHIS)
	results, err := svc.SearchPatients("hospital-a", &model.PatientSearchParams{NationalID: "1111111111111"})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "HN-DISK-FULL", results[0].PatientHN)
}

