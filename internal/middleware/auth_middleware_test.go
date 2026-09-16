package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"agnos-assessment/internal/config"
	"agnos-assessment/internal/middleware"
	"agnos-assessment/internal/model"
	"agnos-assessment/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type dummyStaffRepo struct{}

func (d *dummyStaffRepo) Create(staff *model.Staff) error { return nil }
func (d *dummyStaffRepo) FindByUsernameAndHospital(u, h string) (*model.Staff, error) {
	return nil, nil
}
func (d *dummyStaffRepo) FindByID(id string) (*model.Staff, error) { return nil, nil }

func setupTestRouter() (*gin.Engine, service.AuthService, string) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		JWTSecret:    "middleware-secret-key-12345",
		JWTExpireHrs: 1,
	}
	authSvc := service.NewAuthService(&dummyStaffRepo{}, cfg)

	r := gin.New()
	r.GET("/protected", middleware.AuthMiddleware(authSvc), func(c *gin.Context) {
		hospital, _ := c.Get(middleware.ContextKeyHospital)
		username, _ := c.Get(middleware.ContextKeyUsername)
		c.JSON(http.StatusOK, gin.H{
			"hospital": hospital,
			"username": username,
		})
	})

	return r, authSvc, cfg.JWTSecret
}

func TestAuthMiddleware_ValidToken_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		JWTSecret:    "middleware-secret-key-12345",
		JWTExpireHrs: 1,
	}

	repo := &mockStaffRepoForMiddleware{}
	authSvc := service.NewAuthService(repo, cfg)

	// Create staff and login to obtain real valid token
	createReq := &model.StaffCreateRequest{
		Username: "nurse_somying",
		Password: "Password123!",
		Hospital: "hospital-a",
	}
	_, err := authSvc.CreateStaff(createReq)
	assert.NoError(t, err)

	token, _, err := authSvc.Login(&model.StaffLoginRequest{
		Username: "nurse_somying",
		Password: "Password123!",
		Hospital: "hospital-a",
	})
	assert.NoError(t, err)

	r := gin.New()
	r.GET("/protected", middleware.AuthMiddleware(authSvc), func(c *gin.Context) {
		hospital, _ := c.Get(middleware.ContextKeyHospital)
		username, _ := c.Get(middleware.ContextKeyUsername)
		c.JSON(http.StatusOK, gin.H{
			"hospital": hospital,
			"username": username,
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "hospital-a")
	assert.Contains(t, w.Body.String(), "nurse_somying")
}

type mockStaffRepoForMiddleware struct {
	staff *model.Staff
}

func (m *mockStaffRepoForMiddleware) Create(s *model.Staff) error {
	m.staff = s
	return nil
}
func (m *mockStaffRepoForMiddleware) FindByUsernameAndHospital(u, h string) (*model.Staff, error) {
	if m.staff != nil && m.staff.Username == u && m.staff.Hospital == h {
		return m.staff, nil
	}
	return nil, nil
}
func (m *mockStaffRepoForMiddleware) FindByID(id string) (*model.Staff, error) {
	return m.staff, nil
}

func TestAuthMiddleware_MissingHeader_401(t *testing.T) {
	r, _, _ := setupTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authorization token is missing")
}

func TestAuthMiddleware_MalformedHeader_401(t *testing.T) {
	r, _, _ := setupTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token abcdef12345") // Missing 'Bearer'
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Bearer")
}

func TestAuthMiddleware_InvalidTokenString_401(t *testing.T) {
	r, _, _ := setupTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer bad.jwt.token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired")
}
