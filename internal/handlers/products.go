package handlers

import (
	"hayoon-bite-backend/internal/models"
	"log"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ProductRequest defines the structure for creating/updating a product
type ProductRequest struct {
	Name   string `json:"name" validate:"required"`
	Price  int    `json:"price" validate:"required,gt=0"`
	Recipe []struct {
		InventoryItemID int     `json:"inventory_item_id" validate:"required"`
		QuantityUsed    float64 `json:"quantity_used" validate:"required,gt=0"`
	} `json:"recipe" validate:"required,min=1"`
}

// ProductResponse defines the structure for product responses, including the recipe
type ProductResponse struct {
	ID     uint                 `json:"id"`
	Name   string               `json:"name"`
	Price  int                  `json:"price"`
	Recipe []RecipeItemResponse `json:"recipe"`
}

type RecipeItemResponse struct {
	InventoryItemID   uint    `json:"inventory_item_id"`
	InventoryItemName string  `json:"inventory_item_name"`
	QuantityUsed      float64 `json:"quantity_used"`
	Unit              string  `json:"unit"`
}

// GetProducts handles fetching all products with their recipes
func GetProducts(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var products []models.Product
		// Ambil semua produk
		if err := db.Find(&products).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch products"})
		}

		var response []ProductResponse
		// Untuk setiap produk, ambil resepnya
		for _, p := range products {
			var recipeItems []RecipeItemResponse
			db.Table("recipe_items ri").
				Select("ri.inventory_item_id, ii.name as inventory_item_name, ri.quantity_used, ii.unit").
				Joins("join inventory_items ii on ii.id = ri.inventory_item_id").
				Where("ri.product_id = ?", p.ID).
				Scan(&recipeItems)

			response = append(response, ProductResponse{
				ID:     p.ID,
				Name:   p.Name,
				Price:  int(p.Price),
				Recipe: recipeItems,
			})
		}

		return c.JSON(response)
	}
}

// GetProduct handles fetching a single product by ID
func GetProduct(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid product ID"})
		}

		var product models.Product
		if err := db.First(&product, id).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
		}

		var recipeItems []RecipeItemResponse
		db.Table("recipe_items ri").
			Select("ri.inventory_item_id, ii.name as inventory_item_name, ri.quantity_used, ii.unit").
			Joins("join inventory_items ii on ii.id = ri.inventory_item_id").
			Where("ri.product_id = ?", product.ID).
			Scan(&recipeItems)

		response := ProductResponse{
			ID:     product.ID,
			Name:   product.Name,
			Price:  int(product.Price),
			Recipe: recipeItems,
		}

		return c.JSON(response)
	}
}

// CreateProduct handles creating a new product and its recipe
func CreateProduct(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req ProductRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
		}

		// Gunakan transaction untuk memastikan semua query berhasil atau tidak sama sekali
		err := db.Transaction(func(tx *gorm.DB) error {
			// 1. Buat produk baru
			newProduct := models.Product{
				Name:  req.Name,
				Price: float64(req.Price),
			}
			if err := tx.Create(&newProduct).Error; err != nil {
				return err
			}

			// 2. Buat resepnya
			for _, item := range req.Recipe {
				recipeItem := models.RecipeItem{
					ProductID:       newProduct.ID,
					InventoryItemID: uint(item.InventoryItemID),
					QuantityUsed:    item.QuantityUsed,
				}
				if err := tx.Create(&recipeItem).Error; err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			log.Printf("Error creating product: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create product"})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Product created successfully"})
	}
}

// UpdateProduct handles updating a product and its recipe
func UpdateProduct(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid product ID"})
		}

		var req ProductRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			// 1. Update detail produk
			productUpdate := models.Product{
				Name:  req.Name,
				Price: float64(req.Price),
			}
			if err := tx.Model(&models.Product{}).Where("id = ?", id).Updates(productUpdate).Error; err != nil {
				return err
			}

			// 2. Hapus resep lama
			if err := tx.Where("product_id = ?", id).Delete(&models.RecipeItem{}).Error; err != nil {
				return err
			}

			// 3. Buat resep baru
			for _, item := range req.Recipe {
				recipeItem := models.RecipeItem{
					ProductID:       uint(id),
					InventoryItemID: uint(item.InventoryItemID),
					QuantityUsed:    item.QuantityUsed,
				}
				if err := tx.Create(&recipeItem).Error; err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			log.Printf("Error updating product: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update product"})
		}

		return c.JSON(fiber.Map{"message": "Product updated successfully"})
	}
}

// DeleteProduct handles deleting a product
func DeleteProduct(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid product ID"})
		}

		// GORM akan menghapus produk.
		// ON DELETE CASCADE pada foreign key di tabel recipe_items akan otomatis
		// menghapus resep yang terkait.
		result := db.Delete(&models.Product{}, id)
		if result.Error != nil {
			log.Printf("Error deleting product: %v", result.Error)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete product"})
		}

		if result.RowsAffected == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
		}

		return c.JSON(fiber.Map{"message": "Product deleted successfully"})
	}
}
