package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"hayoon-bite-backend/internal/middleware"
	"hayoon-bite-backend/internal/models"
)

type AuthHandler struct {
	DB *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{DB: db}
}

// Login handles user login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// -----------------------------------------------------
	// >> DEBUG LOG START
	log.Printf("🔍 DEBUG INPUT: Mencoba login dengan user: %s dan password: %s", req.Username, req.Password)
	// -----------------------------------------------------

	// Find user by username
	var user models.User
	if err := h.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Log ini akan muncul jika user "admin" belum ada di DB
			log.Println("❌ DEBUG: User tidak ditemukan di database.")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}
		log.Printf("Database error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	// -----------------------------------------------------
	// >> DEBUG LOG CRUCIAL DATA
	log.Printf("🔍 DEBUG STORED: Hash di DB untuk user %s adalah: %s", user.Username, user.Password)
	if user.Password == "" {
		log.Println("❌ ERROR KRUSIAL: Password hash KOSONG! (Gagal baca dari kolom password_hash)")
	}
	// -----------------------------------------------------

	// Check password
	err := middleware.CheckPassword(req.Password, user.Password)

	// -----------------------------------------------------
	// >> DEBUG LOG COMPARISON RESULT
	if err != nil {
		log.Printf("❌ DEBUG AUTH FAILED: bcrypt comparison GAGAL. Error: %v", err)
	} else {
		log.Println("✅ DEBUG AUTH SUCCESS: Password cocok!")
	}
	// -----------------------------------------------------

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	// Generate JWT token
	token, err := middleware.GenerateJWT(user.ID, user.Role)
	if err != nil {
		log.Printf("Error generating JWT: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error generating authentication token",
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"role":  user.Role,
	})
}

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Username string      `json:"username" validate:"required"`
	Password string      `json:"password" validate:"required,min=6"`
	Role     models.Role `json:"role" validate:"required,oneof=admin karyawan kasir"`
}

// Register handles user registration (for admin only)
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	// Only admin can register new users
	// Ambil data dari JWT Middleware (Locals)
	userRole, ok := c.Locals("userRole").(models.Role)

	// Bypass sementara jika Locals kosong (opsional, untuk dev)
	if !ok {
		// return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Unauthorized access"})
	}

	// Cek Role (hanya Admin yang boleh)
	if ok && userRole != models.RoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only admin can register new users",
		})
	}

	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if username already exists
	existingUser := models.User{}
	if err := h.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username already exists",
		})
	}

	// Hash password
	hashedPassword, err := middleware.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error processing request",
		})
	}

	// Create user
	user := models.User{
		Username: req.Username,
		// PERBAIKAN: Menggunakan field Password
		Password: hashedPassword,
		Role:     req.Role,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		log.Printf("Error creating user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error creating user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"user": fiber.Map{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

// GetProfile returns the current user's profile
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		log.Printf("Database error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	// Don't return password hash
	// PERBAIKAN: Kosongkan field Password sebelum return
	user.Password = ""
	return c.JSON(user)
}
