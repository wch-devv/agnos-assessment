package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Hospital represents a hospital organization
type Hospital struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Code      string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (h *Hospital) BeforeCreate(tx *gorm.DB) (err error) {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return
}

// Staff represents a hospital staff user with login credentials
type Staff struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Username     string         `gorm:"type:varchar(100);not null;uniqueIndex:idx_staff_hospital_username" json:"username"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	Hospital     string         `gorm:"type:varchar(100);not null;uniqueIndex:idx_staff_hospital_username;index" json:"hospital"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (s *Staff) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}

// Patient represents a patient record compatible with Hospital A HIS data structure
type Patient struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Hospital     string         `gorm:"type:varchar(100);not null;uniqueIndex:idx_patient_hospital_hn;index:idx_patient_hospital" json:"hospital"`
	PatientHN    string         `gorm:"type:varchar(50);not null;uniqueIndex:idx_patient_hospital_hn" json:"patient_hn"`
	NationalID   string         `gorm:"type:varchar(20);index:idx_patient_national" json:"national_id,omitempty"`
	PassportID   string         `gorm:"type:varchar(50);index:idx_patient_passport" json:"passport_id,omitempty"`
	FirstNameTH  string         `gorm:"type:varchar(100);not null" json:"first_name_th"`
	MiddleNameTH string         `gorm:"type:varchar(100)" json:"middle_name_th,omitempty"`
	LastNameTH   string         `gorm:"type:varchar(100);not null" json:"last_name_th"`
	FirstNameEN  string         `gorm:"type:varchar(100);not null" json:"first_name_en"`
	MiddleNameEN string         `gorm:"type:varchar(100)" json:"middle_name_en,omitempty"`
	LastNameEN   string         `gorm:"type:varchar(100);not null" json:"last_name_en"`
	DateOfBirth  string         `gorm:"type:date;not null;index" json:"date_of_birth"`
	PhoneNumber  string         `gorm:"type:varchar(30);index" json:"phone_number,omitempty"`
	Email        string         `gorm:"type:varchar(255);index" json:"email,omitempty"`
	Gender       string         `gorm:"type:varchar(2);not null" json:"gender"` // "M", "F"
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (p *Patient) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}

// StaffCreateRequest is the DTO for POST /staff/create
type StaffCreateRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Hospital string `json:"hospital" binding:"required"`
}

// StaffLoginRequest is the DTO for POST /staff/login
type StaffLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Hospital string `json:"hospital" binding:"required"`
}

// PatientSearchParams is the DTO for GET /patient/search (query string) and POST /patient/search (JSON body)
type PatientSearchParams struct {
	NationalID  string `form:"national_id" json:"national_id"`
	PassportID  string `form:"passport_id" json:"passport_id"`
	FirstName   string `form:"first_name" json:"first_name"`
	MiddleName  string `form:"middle_name" json:"middle_name"`
	LastName    string `form:"last_name" json:"last_name"`
	DateOfBirth string `form:"date_of_birth" json:"date_of_birth"`
	PhoneNumber string `form:"phone_number" json:"phone_number"`
	Email       string `form:"email" json:"email"`
}
