package handlers

import (
	"hayoon-bite-backend/internal/models"
	"log"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// InventoryRequest defines the structure for creating/updating an inventory item
type InventoryRequest struct {
	Name       string  `json:"name" validate:"required"`
	StockLevel float64 `json:"stock_level" validate:"gte=0"`
	Unit       string  `json:"unit" validate:"required"`
}

// CreateInventoryItem handles creating a new inventory item
func CreateInventoryItem(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req InventoryRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
		}

		// Cek apakah item dengan nama yang sama sudah ada
		var existing models.InventoryItem
		if err := db.Where("name = ?", req.Name).First(&existing).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Inventory item with this name already exists"})
		}

		newItem := models.InventoryItem{
			Name:       req.Name,
			StockLevel: req.StockLevel,
			Unit:       req.Unit,
		}

		if err := db.Create(&newItem).Error; err != nil {
			log.Printf("Error creating inventory item: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create inventory item"})
		}

		return c.Status(fiber.StatusCreated).JSON(newItem)
	}
}

// UpdateInventoryItem handles updating an existing inventory item
func UpdateInventoryItem(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid item ID"})
		}

		var req InventoryRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
		}

		// Cek duplikasi nama, kecuali untuk item itu sendiri
		var existing models.InventoryItem
		if err := db.Where("name = ? AND id != ?", req.Name, id).First(&existing).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Another inventory item with this name already exists"})
		}

		updateData := models.InventoryItem{
			Name:       req.Name,
			StockLevel: req.StockLevel,
			Unit:       req.Unit,
		}

		result := db.Model(&models.InventoryItem{}).Where("id = ?", id).Updates(updateData)
		if result.Error != nil {
			log.Printf("Error updating inventory item: %v", result.Error)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update inventory item"})
		}

		if result.RowsAffected == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Inventory item not found"})
		}

		return c.JSON(fiber.Map{"message": "Inventory item updated successfully"})
	}
}

// DeleteInventoryItem handles deleting an inventory item
func DeleteInventoryItem(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid item ID"})
		}

		// Cek apakah item ini digunakan di resep
		var recipeCount int64
		db.Model(&models.RecipeItem{}).Where("inventory_item_id = ?", id).Count(&recipeCount)
		if recipeCount > 0 {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Cannot delete item, it is currently used in one or more product recipes."})
		}

		result := db.Delete(&models.InventoryItem{}, id)
		if result.Error != nil {
			log.Printf("Error deleting inventory item: %v", result.Error)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete inventory item"})
		}

		if result.RowsAffected == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Inventory item not found"})
		}

		return c.JSON(fiber.Map{"message": "Inventory item deleted successfully"})
	}
}
