package handlers

import (
	"encoding/json"
	"fmt"
	"hayoon-bite-backend/internal/models"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ProductRequest defines the structure for creating/updating a product
type ProductRequest struct {
	Name   string `json:"name"`
	Price  int    `json:"price"`
	Recipe []struct {
		InventoryItemID int     `json:"inventory_item_id"`
		QuantityUsed    float64 `json:"quantity_used"`
	} `json:"recipe"`
}

// ProductResponse defines the structure for product responses, including the recipe
type ProductResponse struct {
	ID        uint                 `json:"id"`
	Name      string               `json:"name"`
	Price     int                  `json:"price"`
	ImagePath string               `json:"image_path"`
	Recipe    []RecipeItemResponse `json:"recipe"`
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
		// Ambil semua produk, urutkan dari yang terbaru
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
				ID:        p.ID,
				Name:      p.Name,
				Price:     int(p.Price),
				ImagePath: p.ImagePath,
				Recipe:    recipeItems,
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
			ID:        product.ID,
			Name:      product.Name,
			Price:     int(product.Price),
			ImagePath: product.ImagePath,
			Recipe:    recipeItems,
		}

		return c.JSON(response)
	}
}

// CreateProduct handles creating a new product and its recipe
func CreateProduct(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Ambil data form (termasuk file)
		form, err := c.MultipartForm()
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid form data"})
		}

		// Ambil data JSON dari field 'data'
		dataJSON := form.Value["data"]
		if len(dataJSON) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing product data"})
		}

		var req ProductRequest
		if err := json.Unmarshal([]byte(dataJSON[0]), &req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid product JSON data"})
		}

		// Handle file upload
		var imagePath string
		files := form.File["image"]
		if len(files) > 0 {
			file := files[0]
			// Buat nama file unik
			filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), file.Filename)
			savePath := filepath.Join("./public/uploads/products", filename)

			if err := c.SaveFile(file, savePath); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save image"})
			}
			imagePath = "/public/uploads/products/" + filename
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			// 1. Buat produk baru
			newProduct := models.Product{
				Name:      req.Name,
				Price:     float64(req.Price),
				ImagePath: imagePath,
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

		// Ambil data form (termasuk file)
		form, err := c.MultipartForm()
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid form data"})
		}

		// Ambil data JSON dari field 'data'
		dataJSON := form.Value["data"]
		if len(dataJSON) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing product data"})
		}

		var req ProductRequest
		if err := json.Unmarshal([]byte(dataJSON[0]), &req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid product JSON data"})
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			// 1. Ambil produk yang ada
			var existingProduct models.Product
			if err := tx.First(&existingProduct, id).Error; err != nil {
				return fmt.Errorf("product not found")
			}

			// Handle file upload jika ada file baru
			var newImagePath string
			files := form.File["image"]
			if len(files) > 0 {
				file := files[0]
				filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), file.Filename)
				savePath := filepath.Join("./public/uploads/products", filename)

				if err := c.SaveFile(file, savePath); err != nil {
					return fmt.Errorf("failed to save new image")
				}
				newImagePath = "/public/uploads/products/" + filename

				// Hapus gambar lama jika ada
				if existingProduct.ImagePath != "" {
					oldPath := filepath.Join(".", existingProduct.ImagePath)
					os.Remove(oldPath)
				}
				existingProduct.ImagePath = newImagePath
			}

			// 2. Update detail produk
			existingProduct.Name = req.Name
			existingProduct.Price = float64(req.Price)
			if err := tx.Save(&existingProduct).Error; err != nil {
				return err
			}

			// 3. Hapus resep lama
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

		// Ambil data produk untuk menghapus gambar terkait
		var product models.Product
		if err := db.First(&product, id).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
		}

		// Hapus gambar dari server
		if product.ImagePath != "" {
			imagePath := filepath.Join(".", product.ImagePath)
			os.Remove(imagePath)
		}

		result := db.Delete(&product)
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
