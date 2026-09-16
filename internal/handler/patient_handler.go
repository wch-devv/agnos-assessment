package handler

import (
	"net/http"

	"agnos-assessment/internal/middleware"
	"agnos-assessment/internal/model"
	"agnos-assessment/internal/service"

	"github.com/gin-gonic/gin"
)

type PatientHandler struct {
	patientService service.PatientService
}

func NewPatientHandler(patientService service.PatientService) *PatientHandler {
	return &PatientHandler{patientService: patientService}
}

// Search handles GET /patient/search
func (h *PatientHandler) Search(c *gin.Context) {
	// Extract staff hospital from JWT context (strictly enforced)
	hospitalVal, exists := c.Get(middleware.ContextKeyHospital)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Hospital scope not found in authentication context",
		})
		return
	}

	hospital, ok := hospitalVal.(string)
	if !ok || hospital == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Invalid hospital scope in authentication context",
		})
		return
	}

	var params model.PatientSearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid query parameters",
		})
		return
	}

	patients, err := h.patientService.SearchPatients(hospital, &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to search patient records",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(patients),
		"data":   patients,
	})
}
