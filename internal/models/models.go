package models

import "time"

// ==========================================
// INVENTORY & PRODUCT
// ==========================================

type InventoryItem struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"not null;unique" json:"name"`
	StockLevel float64   `gorm:"not null;default:0.00" json:"stock_level"`
	Unit       string    `gorm:"not null" json:"unit"`
	CreatedAt  time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt  time.Time `gorm:"default:now()" json:"updated_at"`
}

type Product struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null;unique" json:"name"`
	Price     float64   `gorm:"not null" json:"price"`
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`
}

type RecipeItem struct {
	ID              uint          `gorm:"primaryKey" json:"id"`
	ProductID       uint          `gorm:"not null" json:"product_id"`
	Product         Product       `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	InventoryItemID uint          `gorm:"not null" json:"inventory_item_id"`
	InventoryItem   InventoryItem `gorm:"foreignKey:InventoryItemID" json:"inventory_item,omitempty"`
	QuantityUsed    float64       `gorm:"not null" json:"quantity_used"`
}

// ==========================================
// POS & TRANSACTIONS
// ==========================================

type Transaction struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	TotalAmount     float64   `gorm:"not null" json:"total_amount"`
	PaymentMethod   string    `gorm:"not null" json:"payment_method"`
	TransactionTime time.Time `gorm:"default:now()" json:"transaction_time"`
}

type TransactionItem struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	TransactionID uint        `gorm:"not null" json:"transaction_id"`
	Transaction   Transaction `gorm:"foreignKey:TransactionID" json:"-"`
	ProductID     uint        `gorm:"not null" json:"product_id"` // TYPO FIXED: gorm.com -> gorm
	Product       Product     `gorm:"foreignKey:ProductID" json:"product"`
	Quantity      int         `gorm:"not null" json:"quantity"`
	Subtotal      float64     `gorm:"not null" json:"subtotal"`
}

// ==========================================
// AUTH & USERS
// ==========================================

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleKaryawan Role = "karyawan"
	RoleKasir    Role = "kasir"
)

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"not null;unique" json:"username"`

	// PERBAIKAN DISINI:
	// Kita namakan fieldnya "Password" (biar enak dipanggil di kodingan)
	// TAPI tag "column:password_hash" memaksa GORM baca dari kolom "password_hash" di DB.
	// Tag json:"-" menyembunyikan password saat data user dikirim ke frontend (SECURITY)
	Password string `gorm:"column:password_hash;not null" json:"-"`

	Role Role `gorm:"type:varchar(20);not null" json:"role"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	Role  Role   `json:"role"`
}
