package repository

import (
	"errors"

	"agnos-assessment/internal/model"

	"gorm.io/gorm"
)

type StaffRepository interface {
	Create(staff *model.Staff) error
	FindByUsernameAndHospital(username, hospital string) (*model.Staff, error)
	FindByID(id string) (*model.Staff, error)
}

type staffRepository struct {
	db *gorm.DB
}

func NewStaffRepository(db *gorm.DB) StaffRepository {
	return &staffRepository{db: db}
}

func (r *staffRepository) Create(staff *model.Staff) error {
	return r.db.Create(staff).Error
}

func (r *staffRepository) FindByUsernameAndHospital(username, hospital string) (*model.Staff, error) {
	var staff model.Staff
	err := r.db.Where("username = ? AND hospital = ?", username, hospital).First(&staff).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &staff, nil
}

func (r *staffRepository) FindByID(id string) (*model.Staff, error) {
	var staff model.Staff
	err := r.db.Where("id = ?", id).First(&staff).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &staff, nil
}
