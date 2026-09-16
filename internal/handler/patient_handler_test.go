package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"agnos-assessment/internal/handler"
	"agnos-assessment/internal/middleware"
	"agnos-assessment/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockPatientService struct {
	patients []model.Patient
	err      error
}

func (m *mockPatientService) SearchPatients(hospital string, params *model.PatientSearchParams) ([]model.Patient, error) {
	if m.err != nil {
		return nil, m.err
	}
	var res []model.Patient
	for _, p := range m.patients {
		if p.Hospital == hospital {
			if params != nil && params.NationalID != "" && p.NationalID != params.NationalID {
				continue
			}
			res = append(res, p)
		}
	}
	return res, nil
}

func TestPatientHandler_Search_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := &mockPatientService{
		patients: []model.Patient{
			{
				ID:          uuid.New(),
				Hospital:    "hospital-a",
				PatientHN:   "HN-A-001",
				NationalID:  "1100501234567",
				FirstNameTH: "สมชาย",
				LastNameTH:  "ใจดี",
				FirstNameEN: "Somchai",
				LastNameEN:  "Jaidee",
				Gender:      "M",
			},
		},
	}
	h := handler.NewPatientHandler(mockSvc)

	r := gin.New()
	r.GET("/patient/search", func(c *gin.Context) {
		// Simulate Auth Middleware injection
		c.Set(middleware.ContextKeyHospital, "hospital-a")
		c.Next()
	}, h.Search)

	req, _ := http.NewRequest(http.MethodGet, "/patient/search?national_id=1100501234567", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Status string          `json:"status"`
		Count  int             `json:"count"`
		Data   []model.Patient `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, 1, resp.Count)
	assert.Equal(t, "HN-A-001", resp.Data[0].PatientHN)
}

func TestPatientHandler_Search_MissingHospitalInContext_401(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := &mockPatientService{}
	h := handler.NewPatientHandler(mockSvc)

	r := gin.New()
	// Router without Auth Middleware (missing ContextKeyHospital)
	r.GET("/patient/search", h.Search)

	req, _ := http.NewRequest(http.MethodGet, "/patient/search", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Hospital scope not found")
}

func TestPatientHandler_Search_PostJSON_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := &mockPatientService{
		patients: []model.Patient{
			{
				ID:          uuid.New(),
				Hospital:    "hospital-a",
				PatientHN:   "HN-A-001",
				NationalID:  "1100501234567",
				FirstNameTH: "สมชาย",
				LastNameTH:  "ใจดี",
				FirstNameEN: "Somchai",
				LastNameEN:  "Jaidee",
				Gender:      "M",
			},
		},
	}
	h := handler.NewPatientHandler(mockSvc)

	r := gin.New()
	r.POST("/patient/search", func(c *gin.Context) {
		c.Set(middleware.ContextKeyHospital, "hospital-a")
		c.Next()
	}, h.Search)

	reqBody := []byte(`{"national_id":"1100501234567"}`)
	req, _ := http.NewRequest(http.MethodPost, "/patient/search", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Status string          `json:"status"`
		Count  int             `json:"count"`
		Data   []model.Patient `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, 1, resp.Count)
	assert.Equal(t, "HN-A-001", resp.Data[0].PatientHN)
}

func TestPatientHandler_Search_EmptyHospitalInContext_401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockPatientService{}
	h := handler.NewPatientHandler(mockSvc)

	r := gin.New()
	r.GET("/patient/search", func(c *gin.Context) {
		c.Set(middleware.ContextKeyHospital, "") // empty hospital
		c.Next()
	}, h.Search)

	req, _ := http.NewRequest(http.MethodGet, "/patient/search", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid hospital scope")
}

func TestPatientHandler_Search_NonStringHospitalInContext_401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockPatientService{}
	h := handler.NewPatientHandler(mockSvc)

	r := gin.New()
	r.GET("/patient/search", func(c *gin.Context) {
		c.Set(middleware.ContextKeyHospital, 12345) // non-string type
		c.Next()
	}, h.Search)

	req, _ := http.NewRequest(http.MethodGet, "/patient/search", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid hospital scope")
}

func TestPatientHandler_Search_InvalidJSON_400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockPatientService{}
	h := handler.NewPatientHandler(mockSvc)

	r := gin.New()
	r.POST("/patient/search", func(c *gin.Context) {
		c.Set(middleware.ContextKeyHospital, "hospital-a")
		c.Next()
	}, h.Search)

	// Invalid malformed JSON
	req, _ := http.NewRequest(http.MethodPost, "/patient/search", bytes.NewBufferString("{invalid-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid JSON request body")
}

func TestPatientHandler_Search_ServiceError_500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockPatientService{
		err: errors.New("db error"),
	}
	h := handler.NewPatientHandler(mockSvc)

	r := gin.New()
	r.GET("/patient/search", func(c *gin.Context) {
		c.Set(middleware.ContextKeyHospital, "hospital-a")
		c.Next()
	}, h.Search)

	req, _ := http.NewRequest(http.MethodGet, "/patient/search", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to search patient records")
}


