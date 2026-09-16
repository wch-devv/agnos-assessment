package repository

import (
	"errors"
	"strings"

	"agnos-assessment/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PatientRepository interface {
	Search(hospital string, params *model.PatientSearchParams) ([]model.Patient, error)
	Create(patient *model.Patient) error
	Upsert(patient *model.Patient) error
	FindByHNAndHospital(hn, hospital string) (*model.Patient, error)
}

type patientRepository struct {
	db *gorm.DB
}

func NewPatientRepository(db *gorm.DB) PatientRepository {
	return &patientRepository{db: db}
}

func (r *patientRepository) Create(patient *model.Patient) error {
	return r.db.Create(patient).Error
}

func (r *patientRepository) Upsert(patient *model.Patient) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "hospital"}, {Name: "patient_hn"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"national_id", "passport_id", "first_name_th", "middle_name_th", "last_name_th",
			"first_name_en", "middle_name_en", "last_name_en", "date_of_birth", "phone_number",
			"email", "gender", "updated_at",
		}),
	}).Create(patient).Error
}

func (r *patientRepository) FindByHNAndHospital(hn, hospital string) (*model.Patient, error) {
	var patient model.Patient
	err := r.db.Where("patient_hn = ? AND hospital = ?", hn, hospital).First(&patient).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &patient, nil
}

// Search queries patients strictly within the staff's hospital
func (r *patientRepository) Search(hospital string, params *model.PatientSearchParams) ([]model.Patient, error) {
	query := r.db.Model(&model.Patient{}).Where("hospital = ?", hospital)

	if params != nil {
		if strings.TrimSpace(params.NationalID) != "" {
			query = query.Where("national_id = ?", strings.TrimSpace(params.NationalID))
		}
		if strings.TrimSpace(params.PassportID) != "" {
			query = query.Where("passport_id = ?", strings.TrimSpace(params.PassportID))
		}
		if strings.TrimSpace(params.FirstName) != "" {
			name := "%" + strings.TrimSpace(params.FirstName) + "%"
			query = query.Where("first_name_th ILIKE ? OR first_name_en ILIKE ?", name, name)
		}
		if strings.TrimSpace(params.MiddleName) != "" {
			mName := "%" + strings.TrimSpace(params.MiddleName) + "%"
			query = query.Where("middle_name_th ILIKE ? OR middle_name_en ILIKE ?", mName, mName)
		}
		if strings.TrimSpace(params.LastName) != "" {
			lName := "%" + strings.TrimSpace(params.LastName) + "%"
			query = query.Where("last_name_th ILIKE ? OR last_name_en ILIKE ?", lName, lName)
		}
		if strings.TrimSpace(params.DateOfBirth) != "" {
			query = query.Where("date_of_birth = ?", strings.TrimSpace(params.DateOfBirth))
		}
		if strings.TrimSpace(params.PhoneNumber) != "" {
			query = query.Where("phone_number = ?", strings.TrimSpace(params.PhoneNumber))
		}
		if strings.TrimSpace(params.Email) != "" {
			query = query.Where("email ILIKE ?", strings.TrimSpace(params.Email))
		}
	}

	var patients []model.Patient
	if err := query.Order("created_at DESC").Find(&patients).Error; err != nil {
		return nil, err
	}

	return patients, nil
}
