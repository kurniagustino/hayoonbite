package handlers

import (
	"time"

	"hayoon-bite-backend/internal/database"
	"hayoon-bite-backend/internal/models"

	"github.com/gofiber/fiber/v2"
)

func GetInventory(c *fiber.Ctx) error {
	var items []models.InventoryItem
	database.DB.Find(&items)
	return c.JSON(items)
}

type StockInRequest struct {
	ID       uint    `json:"id"`
	Quantity float64 `json:"quantity"`
}

func StockIn(c *fiber.Ctx) error {
	var req StockInRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	var item models.InventoryItem
	if err := database.DB.First(&item, req.ID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Inventory item not found"})
	}

	item.StockLevel += req.Quantity
	database.DB.Save(&item)

	return c.JSON(item)
}

type TransactionRequest struct {
	PaymentMethod string `json:"payment_method"`
	Items         []struct {
		ProductID uint `json:"product_id"`
		Quantity  int  `json:"quantity"`
	} `json:"items"`
}

func CreateTransaction(c *fiber.Ctx) error {
	var req TransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to begin transaction"})
	}

	var totalAmount float64
	for _, item := range req.Items {
		var product models.Product
		if err := tx.First(&product, item.ProductID).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
		}
		totalAmount += product.Price * float64(item.Quantity)
	}

	transaction := models.Transaction{
		TotalAmount:   totalAmount,
		PaymentMethod: req.PaymentMethod,
	}
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create transaction"})
	}

	for _, item := range req.Items {
		var product models.Product
		tx.First(&product, item.ProductID)
		transactionItem := models.TransactionItem{
			TransactionID: transaction.ID,
			ProductID:     item.ProductID,
			Quantity:      item.Quantity,
			Subtotal:      product.Price * float64(item.Quantity),
		}
		if err := tx.Create(&transactionItem).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create transaction item: " + err.Error()})
		}

		var recipeItems []models.RecipeItem
		tx.Where("product_id = ?", item.ProductID).Find(&recipeItems)

		for _, recipeItem := range recipeItems {
			var inventoryItem models.InventoryItem
			if err := tx.First(&inventoryItem, recipeItem.InventoryItemID).Error; err != nil {
				tx.Rollback()
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Inventory item not found"})
			}
			inventoryItem.StockLevel -= recipeItem.QuantityUsed * float64(item.Quantity)
			if err := tx.Save(&inventoryItem).Error; err != nil {
				tx.Rollback()
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update inventory"})
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit transaction"})
	}

	return c.JSON(fiber.Map{"message": "Transaction successful"})
}

type FinancialReportResponse struct {
	TotalRevenue   float64 `json:"total_revenue"`
	PaymentMethods []struct {
		PaymentMethod string  `json:"payment_method"`
		TotalAmount   float64 `json:"total_amount"`
	} `json:"payment_methods"`
}

func GetFinancialReport(c *fiber.Ctx) error {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid start_date format. Use YYYY-MM-DD"})
		}
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid end_date format. Use YYYY-MM-DD"})
		}
		endDate = endDate.Add(24*time.Hour - time.Nanosecond)
	}

	var totalRevenue float64
	query := database.DB.Model(&models.Transaction{})
	if !startDate.IsZero() {
		query = query.Where("transaction_time >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("transaction_time <= ?", endDate)
	}
	query.Select("sum(total_amount)").Row().Scan(&totalRevenue)

	var paymentMethods []struct {
		PaymentMethod string
		TotalAmount   float64
	}
	query = database.DB.Model(&models.Transaction{}).Select("payment_method, sum(total_amount) as total_amount").Group("payment_method")
	if !startDate.IsZero() {
		query = query.Where("transaction_time >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("transaction_time <= ?", endDate)
	}
	query.Scan(&paymentMethods)

	response := FinancialReportResponse{
		TotalRevenue: totalRevenue,
	}
	for _, pm := range paymentMethods {
		response.PaymentMethods = append(response.PaymentMethods, struct {
			PaymentMethod string  `json:"payment_method"`
			TotalAmount   float64 `json:"total_amount"`
		}{
			PaymentMethod: pm.PaymentMethod,
			TotalAmount:   pm.TotalAmount,
		})
	}

	return c.JSON(response)
}
