package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath" // Import baru untuk manajemen path file

	"hayoon-bite-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect hanya bertugas untuk KONEKSI saja
func Connect() {
	// ... (Fungsi Connect tidak perlu diubah, tetap sama)
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	sslmode := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("Failed to connect to database. \nError: ", err)
	}

	log.Println("✅ Database connection successful!")
	DB = db
}

// Migrate adalah fungsi KHUSUS untuk menjalankan migrasi dan seeding
func Migrate() {
	// Pastikan DB sudah connect
	if DB == nil {
		Connect()
	}

	// --- 1. JALANKAN MIGRASI SKEMA (CREATE TABLE) DENGAN GORM ---
	log.Println("Running Schema Migrations (Gorm AutoMigrate)...")
	err := DB.AutoMigrate(
		&models.InventoryItem{},
		&models.Product{},
		&models.RecipeItem{},
		&models.Transaction{},
		&models.TransactionItem{},
		&models.User{},
		&models.OperationalCost{}, // <-- TAMBAHKAN MODEL BARU DI SINI
	)

	if err != nil {
		log.Fatal("Schema Migration failed: ", err)
	}
	log.Println("✅ Schema Migrations completed.")

	// --- 2. JALANKAN SEEDING DATA DENGAN RAW SQL ---

	// Tentukan path ke file migrasi Anda
	// Asumsi: File .up.sql ada di folder 'migrations'
	migrationFilePath := filepath.Join("migrations", "000001_initial_schema.up.sql")

	// Baca seluruh konten file SQL
	seederSQL, err := os.ReadFile(migrationFilePath)
	if err != nil {
		log.Fatalf("❌ Failed to read migration/seeder file %s: %v", migrationFilePath, err)
	}

	log.Println("Running Data Seeding (Raw SQL INSERT)...")

	// Eksekusi konten file SQL
	// Karena Gorm AutoMigrate sudah dijalankan, Raw SQL ini akan fokus pada INSERTs
	result := DB.Exec(string(seederSQL))

	if result.Error != nil {
		// Jika ada error pada INSERT (misalnya Foreign Key Error), catat
		log.Fatalf("❌ Data Seeding (INSERT) failed: %v", result.Error)
	}

	// Catat baris yang terpengaruh (opsional)
	log.Printf("Seeding completed. Rows affected: %d\n", result.RowsAffected)
	log.Println("✅ Migrations and Seeding completed successfully!")
}
