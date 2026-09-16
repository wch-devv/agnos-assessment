package main

import (
	"log"

	"agnos-assessment/internal/config"
	"agnos-assessment/internal/handler"
	"agnos-assessment/internal/model"
	"agnos-assessment/internal/repository"
	"agnos-assessment/internal/router"
	"agnos-assessment/internal/service"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func main() {
	log.Println("Starting Hospital Middleware System...")

	// 1. Load Configurations
	cfg := config.LoadConfig()

	// 2. Initialize Database & Auto-Migrate
	db, err := repository.InitDB(cfg)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	// 3. Seed initial demo data for smooth evaluation
	seedDatabase(db)

	// 4. Initialize Repositories
	staffRepo := repository.NewStaffRepository(db)
	patientRepo := repository.NewPatientRepository(db)

	// 5. Initialize Services
	hisClient := service.NewHISClient(cfg.HISBaseURL)
	authService := service.NewAuthService(staffRepo, cfg)
	patientService := service.NewPatientService(patientRepo, hisClient)

	// 6. Initialize Handlers
	staffHandler := handler.NewStaffHandler(authService)
	patientHandler := handler.NewPatientHandler(patientService)

	// 7. Setup Gin Router
	r := router.SetupRouter(staffHandler, patientHandler, authService)

	// 8. Start HTTP Server
	addr := ":" + cfg.Port
	log.Printf("Server is listening on http://localhost%s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}

// seedDatabase seeds initial hospital & patient data for evaluation
func seedDatabase(db *gorm.DB) {
	// 1. Seed Hospitals
	var hospCount int64
	db.Model(&model.Hospital{}).Count(&hospCount)
	if hospCount == 0 {
		hospitals := []model.Hospital{
			{ID: uuid.New(), Code: "hospital-a", Name: "Hospital A Medical Center"},
			{ID: uuid.New(), Code: "hospital-b", Name: "Hospital B Regional Health"},
		}
		db.Create(&hospitals)
		log.Println("Seeded initial hospitals (hospital-a, hospital-b).")
	}

	// 2. Seed Sample Patients for Hospital A and Hospital B
	var patientCount int64
	db.Model(&model.Patient{}).Count(&patientCount)
	if patientCount == 0 {
		patients := []model.Patient{
			{
				ID:           uuid.New(),
				Hospital:     "hospital-a",
				PatientHN:    "HN-10001",
				NationalID:   "1100501234567",
				PassportID:   "AA1234567",
				FirstNameTH:  "สมชาย",
				MiddleNameTH: "",
				LastNameTH:   "ใจดี",
				FirstNameEN:  "Somchai",
				MiddleNameEN: "",
				LastNameEN:   "Jaidee",
				DateOfBirth:  "1990-05-15",
				PhoneNumber:  "0812345678",
				Email:        "somchai@example.com",
				Gender:       "M",
			},
			{
				ID:           uuid.New(),
				Hospital:     "hospital-a",
				PatientHN:    "HN-10002",
				NationalID:   "1100507654321",
				PassportID:   "BB9876543",
				FirstNameTH:  "วิภาดา",
				MiddleNameTH: "",
				LastNameTH:   "รักสงบ",
				FirstNameEN:  "Wiphada",
				MiddleNameEN: "",
				LastNameEN:   "Raksangob",
				DateOfBirth:  "1995-10-20",
				PhoneNumber:  "0898765432",
				Email:        "wiphada@example.com",
				Gender:       "F",
			},
			{
				ID:           uuid.New(),
				Hospital:     "hospital-b",
				PatientHN:    "HN-20001",
				NationalID:   "1200901234567",
				PassportID:   "CC5556667",
				FirstNameTH:  "อนันต์",
				MiddleNameTH: "",
				LastNameTH:   "สุขเจริญ",
				FirstNameEN:  "Anan",
				MiddleNameEN: "",
				LastNameEN:   "Sukcharoen",
				DateOfBirth:  "1988-02-14",
				PhoneNumber:  "0855551234",
				Email:        "anan@example.com",
				Gender:       "M",
			},
		}
		db.Create(&patients)
		log.Println("Seeded sample patients for hospital-a and hospital-b.")
	}
}
