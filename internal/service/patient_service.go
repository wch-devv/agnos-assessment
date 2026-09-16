package service

import (
	"log"
	"strings"

	"agnos-assessment/internal/model"
	"agnos-assessment/internal/repository"
)

type PatientService interface {
	SearchPatients(hospital string, params *model.PatientSearchParams) ([]model.Patient, error)
}

type patientService struct {
	patientRepo repository.PatientRepository
	hisClient   HISClient
}

func NewPatientService(patientRepo repository.PatientRepository, hisClient HISClient) PatientService {
	return &patientService{
		patientRepo: patientRepo,
		hisClient:   hisClient,
	}
}

func (s *patientService) SearchPatients(hospital string, params *model.PatientSearchParams) ([]model.Patient, error) {
	// 1. Search in local database strictly scoped by staff's hospital
	patients, err := s.patientRepo.Search(hospital, params)
	if err != nil {
		return nil, err
	}

	// 2. If results found, return them
	if len(patients) > 0 {
		return patients, nil
	}

	// 3. Middleware HIS Integration (Hospital A)
	// If the querying staff is from Hospital A and searched by national_id or passport_id,
	// query the external Hospital A HIS API
	if strings.EqualFold(hospital, "hospital-a") && params != nil {
		searchID := strings.TrimSpace(params.NationalID)
		if searchID == "" {
			searchID = strings.TrimSpace(params.PassportID)
		}

		if searchID != "" && s.hisClient != nil {
			hisPatient, err := s.hisClient.FetchPatientFromHospitalA(searchID)
			if err != nil {
				// Log error but do not break search, fallback gracefully
				log.Printf("Warning: HIS Hospital A lookup failed for ID %s: %v", searchID, err)
			} else if hisPatient != nil {
				// Enforce hospital assignment
				hisPatient.Hospital = "hospital-a"
				// Upsert into local database
				if saveErr := s.patientRepo.Upsert(hisPatient); saveErr != nil {
					log.Printf("Warning: Failed to cache HIS patient to local DB: %v", saveErr)
				}
				return []model.Patient{*hisPatient}, nil
			}
		}
	}

	return []model.Patient{}, nil
}
