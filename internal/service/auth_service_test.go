package service_test

import (
	"errors"
	"strings"
	"testing"

	"agnos-assessment/internal/config"
	"agnos-assessment/internal/model"
	"agnos-assessment/internal/service"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// MockStaffRepository for unit testing
type mockStaffRepository struct {
	staffs    map[string]*model.Staff
	findErr   error
	createErr error
}

func newMockStaffRepository() *mockStaffRepository {
	return &mockStaffRepository{
		staffs: make(map[string]*model.Staff),
	}
}

func (m *mockStaffRepository) key(username, hospital string) string {
	return hospital + ":" + username
}

func (m *mockStaffRepository) Create(staff *model.Staff) error {
	if m.createErr != nil {
		return m.createErr
	}
	k := m.key(staff.Username, staff.Hospital)
	if _, exists := m.staffs[k]; exists {
		return errors.New("duplicate key violation")
	}
	m.staffs[k] = staff
	return nil
}

func (m *mockStaffRepository) FindByUsernameAndHospital(username, hospital string) (*model.Staff, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	k := m.key(username, hospital)
	if staff, ok := m.staffs[k]; ok {
		return staff, nil
	}
	return nil, nil
}

func (m *mockStaffRepository) FindByID(id string) (*model.Staff, error) {
	for _, staff := range m.staffs {
		if staff.ID.String() == id {
			return staff, nil
		}
	}
	return nil, nil
}

func setupAuthService() (service.AuthService, *mockStaffRepository) {
	cfg := &config.Config{
		JWTSecret:    "test-secret-key-12345",
		JWTExpireHrs: 1,
	}
	repo := newMockStaffRepository()
	return service.NewAuthService(repo, cfg), repo
}

// =========================================================================
// Positive Test Scenarios
// =========================================================================

func TestAuthService_CreateStaff_Success(t *testing.T) {
	svc, _ := setupAuthService()

	req := &model.StaffCreateRequest{
		Username: "doctor_somchai",
		Password: "SecurePassword123!",
		Hospital: "hospital-a",
	}

	staff, err := svc.CreateStaff(req)
	assert.NoError(t, err)
	assert.NotNil(t, staff)
	assert.Equal(t, "doctor_somchai", staff.Username)
	assert.Equal(t, "hospital-a", staff.Hospital)
	assert.NotEqual(t, "SecurePassword123!", staff.PasswordHash, "Password must be hashed")
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(staff.PasswordHash), []byte("SecurePassword123!")))
}

func TestAuthService_Login_Success(t *testing.T) {
	svc, _ := setupAuthService()

	// 1. Create staff first
	createReq := &model.StaffCreateRequest{
		Username: "nurse_joy",
		Password: "JoyPassword123!",
		Hospital: "hospital-a",
	}
	_, err := svc.CreateStaff(createReq)
	assert.NoError(t, err)

	// 2. Login with correct credentials
	loginReq := &model.StaffLoginRequest{
		Username: "nurse_joy",
		Password: "JoyPassword123!",
		Hospital: "hospital-a",
	}
	token, staff, err := svc.Login(loginReq)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotNil(t, staff)
	assert.Equal(t, "nurse_joy", staff.Username)
	assert.Equal(t, "hospital-a", staff.Hospital)

	// 3. Validate returned JWT token
	claims, err := svc.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, "nurse_joy", claims.Username)
	assert.Equal(t, "hospital-a", claims.Hospital)
}

// =========================================================================
// Negative Test Scenarios
// =========================================================================

func TestAuthService_CreateStaff_DuplicateUsernameInSameHospital(t *testing.T) {
	svc, _ := setupAuthService()

	req := &model.StaffCreateRequest{
		Username: "doctor_somchai",
		Password: "Password123!",
		Hospital: "hospital-a",
	}

	// First creation succeeds
	_, err := svc.CreateStaff(req)
	assert.NoError(t, err)

	// Duplicate creation in same hospital must fail
	_, err = svc.CreateStaff(req)
	assert.Error(t, err)
	assert.ErrorIs(t, err, service.ErrStaffAlreadyExists)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	svc, _ := setupAuthService()

	createReq := &model.StaffCreateRequest{
		Username: "user1",
		Password: "CorrectPassword123!",
		Hospital: "hospital-a",
	}
	_, err := svc.CreateStaff(createReq)
	assert.NoError(t, err)

	loginReq := &model.StaffLoginRequest{
		Username: "user1",
		Password: "WrongPassword!",
		Hospital: "hospital-a",
	}
	token, staff, err := svc.Login(loginReq)
	assert.Error(t, err)
	assert.ErrorIs(t, err, service.ErrInvalidCredentials)
	assert.Empty(t, token)
	assert.Nil(t, staff)
}

func TestAuthService_Login_WrongHospital(t *testing.T) {
	svc, _ := setupAuthService()

	createReq := &model.StaffCreateRequest{
		Username: "hospital_a_staff",
		Password: "Password123!",
		Hospital: "hospital-a",
	}
	_, err := svc.CreateStaff(createReq)
	assert.NoError(t, err)

	// Attempt login with hospital-b
	loginReq := &model.StaffLoginRequest{
		Username: "hospital_a_staff",
		Password: "Password123!",
		Hospital: "hospital-b",
	}
	token, staff, err := svc.Login(loginReq)
	assert.Error(t, err)
	assert.ErrorIs(t, err, service.ErrInvalidCredentials)
	assert.Empty(t, token)
	assert.Nil(t, staff)
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	svc, _ := setupAuthService()

	claims, err := svc.ValidateToken("invalid.token.string")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestAuthService_AllowsSameUsernameInDifferentHospitals(t *testing.T) {
	svc, _ := setupAuthService()

	reqA := &model.StaffCreateRequest{
		Username: "admin",
		Password: "PasswordA123!",
		Hospital: "hospital-a",
	}
	staffA, err := svc.CreateStaff(reqA)
	assert.NoError(t, err)
	assert.NotNil(t, staffA)

	// Same username "admin" in hospital-b should succeed
	reqB := &model.StaffCreateRequest{
		Username: "admin",
		Password: "PasswordB123!",
		Hospital: "hospital-b",
	}
	staffB, err := svc.CreateStaff(reqB)
	assert.NoError(t, err)
	assert.NotNil(t, staffB)
	assert.NotEqual(t, staffA.ID, staffB.ID)
}

func TestAuthService_CreateStaff_RepoFindError(t *testing.T) {
	svc, repo := setupAuthService()
	repo.findErr = errors.New("db error on find")

	req := &model.StaffCreateRequest{
		Username: "staff_fail",
		Password: "Password123!",
		Hospital: "hospital-a",
	}
	staff, err := svc.CreateStaff(req)
	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.Contains(t, err.Error(), "failed to query staff")
}

func TestAuthService_CreateStaff_RepoCreateError(t *testing.T) {
	svc, repo := setupAuthService()
	repo.createErr = errors.New("db error on insert")

	req := &model.StaffCreateRequest{
		Username: "staff_fail2",
		Password: "Password123!",
		Hospital: "hospital-a",
	}
	staff, err := svc.CreateStaff(req)
	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.Contains(t, err.Error(), "failed to persist staff")
}

func TestAuthService_Login_RepoFindError(t *testing.T) {
	svc, repo := setupAuthService()
	repo.findErr = errors.New("db connection lost")

	req := &model.StaffLoginRequest{
		Username: "staff1",
		Password: "Password123!",
		Hospital: "hospital-a",
	}
	token, staff, err := svc.Login(req)
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Nil(t, staff)
	assert.Contains(t, err.Error(), "database error")
}

func TestAuthService_ValidateToken_WrongSigningMethod(t *testing.T) {
	svc, _ := setupAuthService()

	// Token signed with 'none' alg
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"username": "hacker",
	})
	tokenStr, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	claims, err := svc.ValidateToken(tokenStr)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "unexpected signing method")
}

func TestAuthService_CreateStaff_PasswordTooLong(t *testing.T) {
	svc, _ := setupAuthService()
	longPassword := strings.Repeat("A", 100)
	req := &model.StaffCreateRequest{
		Username: "staff_long",
		Password: longPassword,
		Hospital: "hospital-a",
	}
	staff, err := svc.CreateStaff(req)
	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.Contains(t, err.Error(), "failed to hash password")
}


