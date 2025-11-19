package handlers

import (
	"hayoon-bite-backend/internal/middleware"
	"hayoon-bite-backend/internal/models"
	"log"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// UserResponse defines the structure for user data sent to the client
type UserResponse struct {
	ID       uint        `json:"id"`
	Username string      `json:"username"`
	Role     models.Role `json:"role"`
}

// UpdateUserRequest defines the structure for updating a user
type UpdateUserRequest struct {
	Username string      `json:"username" validate:"required"`
	Password string      `json:"password,omitempty"` // Password is optional
	Role     models.Role `json:"role" validate:"required,oneof=admin karyawan kasir"`
}

// GetUsers handles fetching all users
func GetUsers(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var users []models.User
		if err := db.Find(&users).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch users"})
		}

		// Transform users to UserResponse to avoid sending password hash
		var response []UserResponse
		for _, user := range users {
			response = append(response, UserResponse{
				ID:       user.ID,
				Username: user.Username,
				Role:     user.Role,
			})
		}

		return c.JSON(response)
	}
}

// UpdateUser handles updating a user's details
func UpdateUser(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
		}

		var req UpdateUserRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
		}

		// Find the user to update
		var user models.User
		if err := db.First(&user, id).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
		}

		// Update fields
		user.Username = req.Username
		user.Role = req.Role

		// If a new password is provided, hash and update it
		if req.Password != "" {
			hashedPassword, err := middleware.HashPassword(req.Password)
			if err != nil {
				log.Printf("Error hashing password: %v", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error processing password"})
			}
			user.Password = hashedPassword
		}

		if err := db.Save(&user).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user"})
		}

		return c.JSON(fiber.Map{"message": "User updated successfully"})
	}
}

// DeleteUser handles deleting a user
func DeleteUser(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
		}

		result := db.Delete(&models.User{}, id)
		if result.Error != nil || result.RowsAffected == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found or could not be deleted"})
		}

		return c.JSON(fiber.Map{"message": "User deleted successfully"})
	}
}
