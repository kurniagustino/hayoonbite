package handlers

import (
	"hayoon-bite-backend/internal/models"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// OperationalCostRequest mendefinisikan struktur untuk membuat/memperbarui biaya operasional
type OperationalCostRequest struct {
	Description string    `json:"description" validate:"required"`
	Amount      float64   `json:"amount" validate:"required,gt=0"`
	Category    string    `json:"category"`
	Date        time.Time `json:"date" validate:"required"`
}

// CreateOperationalCost menangani pembuatan entri biaya operasional baru
func CreateOperationalCost(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req OperationalCostRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
		}

		newCost := models.OperationalCost{
			Description: req.Description,
			Amount:      req.Amount,
			Category:    req.Category,
			Date:        req.Date,
		}

		if err := db.Create(&newCost).Error; err != nil {
			log.Printf("Error creating operational cost: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create operational cost"})
		}

		return c.Status(fiber.StatusCreated).JSON(newCost)
	}
}

// GetOperationalCosts menangani pengambilan daftar biaya operasional
func GetOperationalCosts(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var costs []models.OperationalCost
		// Urutkan dari yang terbaru
		if err := db.Order("date desc").Find(&costs).Error; err != nil {
			log.Printf("Error fetching operational costs: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch operational costs"})
		}
		return c.JSON(costs)
	}
}

// UpdateOperationalCost menangani pembaruan entri biaya operasional
func UpdateOperationalCost(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid cost ID"})
		}

		var req OperationalCostRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
		}

		var existingCost models.OperationalCost
		if err := db.First(&existingCost, id).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Operational cost not found"})
		}

		existingCost.Description = req.Description
		existingCost.Amount = req.Amount
		existingCost.Category = req.Category
		existingCost.Date = req.Date

		if err := db.Save(&existingCost).Error; err != nil {
			log.Printf("Error updating operational cost: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update operational cost"})
		}

		return c.JSON(fiber.Map{"message": "Operational cost updated successfully", "data": existingCost})
	}
}

// DeleteOperationalCost menangani penghapusan entri biaya operasional
func DeleteOperationalCost(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid cost ID"})
		}

		result := db.Delete(&models.OperationalCost{}, id)
		if result.Error != nil {
			log.Printf("Error deleting operational cost: %v", result.Error)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete operational cost"})
		}

		if result.RowsAffected == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Operational cost not found"})
		}

		return c.JSON(fiber.Map{"message": "Operational cost deleted successfully"})
	}
}
