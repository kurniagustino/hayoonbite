package main

import (
	"log"

	"hayoon-bite-backend/internal/database"
	"hayoon-bite-backend/internal/handlers"
	"hayoon-bite-backend/internal/middleware"
	"hayoon-bite-backend/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
)

func main() {
	// 1. LOAD .ENV DULUAN!
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: File .env tidak ditemukan, menggunakan environment system (jika ada)")
	}

	// 2. Connect Database
	database.Connect()

	// ---------------------------------------------------------
	// 3. SETUP TEMPLATE ENGINE
	// ---------------------------------------------------------
	engine := html.New("./views", ".html")
	engine.Reload(true) // Auto reload html saat dev

	// --- DAFTARKAN FUNGSI CUSTOM DI SINI ---

	// A. Fungsi "default" (untuk menangani nilai kosong)
	engine.AddFunc("default", func(d interface{}, s string) interface{} {
		if s != "" {
			return s
		}
		return d
	})

	// B. FUNGSI BARU: "activeClass" (PENTING UNTUK SIDEBAR)
	// Fungsi ini mengecek apakah menu sedang aktif, lalu memberi warna oranye
	engine.AddFunc("activeClass", func(currentTitle, menuTitle string) string {
		if currentTitle == menuTitle {
			return "bg-brand-orange text-brand-dark font-semibold"
		}
		return ""
	})

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Use(logger.New())

	// ---------------------------------------------------------
	// 4. STATIC FILES (PENTING UTK TAILWIND)
	// ---------------------------------------------------------
	app.Static("/public", "./public")
	app.Static("/public/uploads", "./public/uploads")

	// ---------------------------------------------------------
	// 5. ROUTE HALAMAN WEB (Render HTML)
	// ---------------------------------------------------------
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title": "Hayoon Bite - Premium Toast",
		})
	})

	// DASHBOARD
	app.Get("/admin", func(c *fiber.Ctx) error {
		return c.Render("admin/dashboard", fiber.Map{
			"Title":           "Admin Dashboard",
			"PageTitle":       "Dashboard",
			"PageDescription": "Ringkasan laporan keuangan dan inventory",
		}, "layouts/admin")
	})

	// PRODUK
	app.Get("/admin/products", func(c *fiber.Ctx) error {
		return c.Render("admin/products", fiber.Map{
			"Title":           "Manajemen Produk",
			"PageTitle":       "Produk & Resep",
			"PageDescription": "Kelola daftar produk dan resep yang dijual",
		}, "layouts/admin")
	})

	// INVENTARIS
	app.Get("/admin/inventory", func(c *fiber.Ctx) error {
		return c.Render("admin/inventory", fiber.Map{
			"Title":           "Manajemen Inventaris",
			"PageTitle":       "Inventaris",
			"PageDescription": "Kelola daftar bahan baku dan stok",
		}, "layouts/admin")
	})

	// PENGGUNA
	app.Get("/admin/users", func(c *fiber.Ctx) error {
		return c.Render("admin/users", fiber.Map{
			"Title":           "Manajemen Pengguna",
			"PageTitle":       "Pengguna",
			"PageDescription": "Kelola akun pengguna dan hak akses",
		}, "layouts/admin")
	})

	// BIAYA OPERASIONAL
	app.Get("/admin/operational-costs", func(c *fiber.Ctx) error {
		return c.Render("admin/operational_costs", fiber.Map{
			"Title":           "Biaya Operasional",
			"PageTitle":       "Biaya Operasional",
			"PageDescription": "Kelola pengeluaran di luar bahan baku",
		}, "layouts/admin")
	})

	// ---------------------------------------------------------
	// 6. API ENDPOINTS
	// ---------------------------------------------------------

	// Inisialisasi Handler Auth
	authHandler := handlers.NewAuthHandler(database.DB)

	api := app.Group("/api/v1")

	// === PUBLIC ROUTES ===
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "Running", "message": "API Ready"})
	})
	api.Post("/login", authHandler.Login)

	// === PROTECTED ROUTES (JWT) ===
	api.Use(middleware.JWTProtected())

	// User Profile
	api.Get("/me", authHandler.GetProfile)

	// Admin Routes
	admin := api.Group("/admin")
	admin.Use(middleware.RoleProtected(models.RoleAdmin))
	admin.Post("/register", authHandler.Register)
	admin.Get("/users", handlers.GetUsers(database.DB))
	admin.Put("/users/:id", handlers.UpdateUser(database.DB))
	admin.Delete("/users/:id", handlers.DeleteUser(database.DB))

	// Inventory Routes
	inventory := api.Group("/inventory")
	inventory.Get("", handlers.GetInventory)
	inventory.Post("/stock-in", handlers.StockIn)
	inventory.Post("", handlers.CreateInventoryItem(database.DB))
	inventory.Put("/:id", handlers.UpdateInventoryItem(database.DB))
	inventory.Delete("/:id", handlers.DeleteInventoryItem(database.DB))

	// Product Routes (Admin)
	products := api.Group("/products")
	products.Get("", handlers.GetProducts(database.DB))
	products.Post("", handlers.CreateProduct(database.DB))
	products.Get("/:id", handlers.GetProduct(database.DB))
	products.Put("/:id", handlers.UpdateProduct(database.DB))
	products.Delete("/:id", handlers.DeleteProduct(database.DB))

	// Operational Costs Routes (Admin)
	// PASTIKAN FILE handlers/operational_costs.go SUDAH DIBUAT
	// JIKA BELUM, BAGIAN INI AKAN ERROR DI TERMINAL
	opCosts := api.Group("/operational-costs")
	opCosts.Get("", handlers.GetOperationalCosts(database.DB))
	opCosts.Post("", handlers.CreateOperationalCost(database.DB))
	opCosts.Put("/:id", handlers.UpdateOperationalCost(database.DB))
	opCosts.Delete("/:id", handlers.DeleteOperationalCost(database.DB))

	// POS Routes
	pos := api.Group("/pos")
	pos.Use(middleware.RoleProtected(models.RoleKasir, models.RoleAdmin))
	pos.Post("/transactions", handlers.CreateTransaction)

	// Reports Routes
	reports := api.Group("/reports")
	reports.Use(middleware.RoleProtected(models.RoleAdmin, models.RoleKaryawan))
	reports.Get("/financial", handlers.GetFinancialReport)

	log.Println("Server berjalan di port :8080")
	log.Fatal(app.Listen(":8080"))
}
