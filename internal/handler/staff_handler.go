package handler

import (
	"errors"
	"net/http"

	"agnos-assessment/internal/model"
	"agnos-assessment/internal/service"

	"github.com/gin-gonic/gin"
)

type StaffHandler struct {
	authService service.AuthService
}

func NewStaffHandler(authService service.AuthService) *StaffHandler {
	return &StaffHandler{authService: authService}
}

// CreateStaff handles POST /staff/create
func (h *StaffHandler) CreateStaff(c *gin.Context) {
	var req model.StaffCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body: username, password (min 6 chars), and hospital are required",
		})
		return
	}

	staff, err := h.authService.CreateStaff(&req)
	if err != nil {
		if errors.Is(err, service.ErrStaffAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{
				"status":  "error",
				"message": "Username already exists in this hospital",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to create staff member",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Staff member created successfully",
		"data": gin.H{
			"id":         staff.ID,
			"username":   staff.Username,
			"hospital":   staff.Hospital,
			"created_at": staff.CreatedAt,
		},
	})
}

// Login handles POST /staff/login
func (h *StaffHandler) Login(c *gin.Context) {
	var req model.StaffLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Username, password, and hospital are required",
		})
		return
	}

	token, staff, err := h.authService.Login(&req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid username, password, or hospital",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Authentication failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Login successful",
		"data": gin.H{
			"token":      token,
			"token_type": "Bearer",
			"staff": gin.H{
				"id":       staff.ID,
				"username": staff.Username,
				"hospital": staff.Hospital,
			},
		},
	})
}
