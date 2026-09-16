package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"agnos-assessment/internal/handler"
	"agnos-assessment/internal/model"
	"agnos-assessment/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockAuthServiceForHandler struct {
	staffs    map[string]string // username -> password
	createErr error
	loginErr  error
}

func (m *mockAuthServiceForHandler) CreateStaff(req *model.StaffCreateRequest) (*model.Staff, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	if _, ok := m.staffs[req.Username]; ok {
		return nil, service.ErrStaffAlreadyExists
	}
	m.staffs[req.Username] = req.Password
	return &model.Staff{
		ID:       uuid.New(),
		Username: req.Username,
		Hospital: req.Hospital,
	}, nil
}

func (m *mockAuthServiceForHandler) Login(req *model.StaffLoginRequest) (string, *model.Staff, error) {
	if m.loginErr != nil {
		return "", nil, m.loginErr
	}
	pwd, ok := m.staffs[req.Username]
	if !ok || pwd != req.Password {
		return "", nil, service.ErrInvalidCredentials
	}
	return "mock-jwt-token-12345", &model.Staff{
		ID:       uuid.New(),
		Username: req.Username,
		Hospital: req.Hospital,
	}, nil
}

func (m *mockAuthServiceForHandler) ValidateToken(tokenStr string) (*service.JWTClaims, error) {
	if tokenStr == "mock-jwt-token-12345" {
		return &service.JWTClaims{
			StaffID:  uuid.New().String(),
			Username: "testuser",
			Hospital: "hospital-a",
		}, nil
	}
	return nil, service.ErrInvalidCredentials
}

func TestStaffHandler_CreateStaff_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := &mockAuthServiceForHandler{staffs: make(map[string]string)}
	h := handler.NewStaffHandler(mockAuth)

	r := gin.New()
	r.POST("/staff/create", h.CreateStaff)

	body, _ := json.Marshal(model.StaffCreateRequest{
		Username: "doctor_win",
		Password: "Password123!",
		Hospital: "hospital-a",
	})
	req, _ := http.NewRequest(http.MethodPost, "/staff/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Staff member created successfully")
}

func TestStaffHandler_CreateStaff_Duplicate_409(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := &mockAuthServiceForHandler{staffs: map[string]string{"doctor_win": "Password123!"}}
	h := handler.NewStaffHandler(mockAuth)

	r := gin.New()
	r.POST("/staff/create", h.CreateStaff)

	body, _ := json.Marshal(model.StaffCreateRequest{
		Username: "doctor_win",
		Password: "Password123!",
		Hospital: "hospital-a",
	})
	req, _ := http.NewRequest(http.MethodPost, "/staff/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "Username already exists")
}

func TestStaffHandler_Login_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := &mockAuthServiceForHandler{staffs: map[string]string{"doctor_win": "Password123!"}}
	h := handler.NewStaffHandler(mockAuth)

	r := gin.New()
	r.POST("/staff/login", h.Login)

	body, _ := json.Marshal(model.StaffLoginRequest{
		Username: "doctor_win",
		Password: "Password123!",
		Hospital: "hospital-a",
	})
	req, _ := http.NewRequest(http.MethodPost, "/staff/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "mock-jwt-token-12345")
}

func TestStaffHandler_Login_WrongCredentials_401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := &mockAuthServiceForHandler{staffs: map[string]string{"doctor_win": "Password123!"}}
	h := handler.NewStaffHandler(mockAuth)

	r := gin.New()
	r.POST("/staff/login", h.Login)

	body, _ := json.Marshal(model.StaffLoginRequest{
		Username: "doctor_win",
		Password: "WrongPassword!",
		Hospital: "hospital-a",
	})
	req, _ := http.NewRequest(http.MethodPost, "/staff/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestStaffHandler_CreateStaff_InvalidBody_400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := &mockAuthServiceForHandler{staffs: make(map[string]string)}
	h := handler.NewStaffHandler(mockAuth)

	r := gin.New()
	r.POST("/staff/create", h.CreateStaff)

	req, _ := http.NewRequest(http.MethodPost, "/staff/create", bytes.NewBufferString("{bad-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")
}

func TestStaffHandler_CreateStaff_InternalError_500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := &mockAuthServiceForHandler{
		staffs:    make(map[string]string),
		createErr: errors.New("unexpected db failure"),
	}
	h := handler.NewStaffHandler(mockAuth)

	r := gin.New()
	r.POST("/staff/create", h.CreateStaff)

	body, _ := json.Marshal(model.StaffCreateRequest{
		Username: "doctor_win",
		Password: "Password123!",
		Hospital: "hospital-a",
	})
	req, _ := http.NewRequest(http.MethodPost, "/staff/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to create staff member")
}

func TestStaffHandler_Login_InvalidBody_400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := &mockAuthServiceForHandler{staffs: make(map[string]string)}
	h := handler.NewStaffHandler(mockAuth)

	r := gin.New()
	r.POST("/staff/login", h.Login)

	req, _ := http.NewRequest(http.MethodPost, "/staff/login", bytes.NewBufferString("{bad-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Username, password, and hospital are required")
}

func TestStaffHandler_Login_InternalError_500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := &mockAuthServiceForHandler{
		staffs:   map[string]string{"doctor_win": "Password123!"},
		loginErr: errors.New("db down"),
	}
	h := handler.NewStaffHandler(mockAuth)

	r := gin.New()
	r.POST("/staff/login", h.Login)

	body, _ := json.Marshal(model.StaffLoginRequest{
		Username: "doctor_win",
		Password: "Password123!",
		Hospital: "hospital-a",
	})
	req, _ := http.NewRequest(http.MethodPost, "/staff/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Authentication failed")
}

