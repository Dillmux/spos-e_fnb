package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SystemConfig menyimpan konfigurasi runtime aplikasi dan status inisialisasi setup.
type SystemConfig struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Key       string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"key"`
	Value     string    `gorm:"type:text;not null" json:"value"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

// Outlet menyimpan metadata outlet, mode layanan, konfigurasi pajak, service charge, dan footer struk.
type Outlet struct {
	ID                string    `gorm:"type:char(36);primaryKey" json:"id"`
	Name              string    `gorm:"type:varchar(150);not null" json:"name"`
	Tagline           string    `gorm:"type:varchar(255)" json:"tagline"`
	Address           string    `gorm:"type:text;not null" json:"address"`
	Phone             string    `gorm:"type:varchar(20);not null" json:"phone"`
	LogoURL           string    `gorm:"type:varchar(255)" json:"logo_url"`
	BusinessType      string    `gorm:"type:varchar(30);not null" json:"business_type"`  // cafe, resto, cloud_kitchen
	ServiceModes      string    `gorm:"type:text;not null" json:"service_modes"`          // JSON string array e.g. ["dine_in","take_away"]
	TaxPercentage     float64   `gorm:"type:decimal(5,2);default:0.00" json:"tax_percentage"`
	ServicePercentage float64   `gorm:"type:decimal(5,2);default:0.00" json:"service_percentage"`
	IsTaxInclusive    bool      `gorm:"default:false" json:"is_tax_inclusive"`
	ReceiptFooter     string    `gorm:"type:text" json:"receipt_footer"`
	CreatedAt         time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt         time.Time `gorm:"not null" json:"updated_at"`
}

func (o *Outlet) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}

// User menyimpan kredensial autentikasi dan peran (superadmin, manager, cashier, kitchen).
type User struct {
	ID           string    `gorm:"type:char(36);primaryKey" json:"id"`
	OutletID     string    `gorm:"type:char(36);not null;index" json:"outlet_id"`
	Name         string    `gorm:"type:varchar(100);not null" json:"name"`
	Username     string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"type:varchar(100)" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	AuthPINHash  string    `gorm:"type:varchar(255)" json:"-"`
	Role         string    `gorm:"type:varchar(20);not null" json:"role"` // superadmin, manager, cashier, kitchen
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null" json:"updated_at"`

	Outlet *Outlet `gorm:"foreignKey:OutletID;references:ID;constraint:OnDelete:RESTRICT" json:"outlet,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// Shift mencatat sesi shift laci kasir dan rekonsiliasi saldo kasir.
type Shift struct {
	ID             string     `gorm:"type:char(36);primaryKey" json:"id"`
	UserID         string     `gorm:"type:char(36);not null;index:idx_shifts_user_status" json:"user_id"`
	StartTime      time.Time  `gorm:"not null" json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
	OpeningCash    float64    `gorm:"type:decimal(12,2);not null" json:"opening_cash"`
	ExpectedCash   *float64   `gorm:"type:decimal(12,2)" json:"expected_cash"`
	ActualCash     *float64   `gorm:"type:decimal(12,2)" json:"actual_cash"`
	CashDifference *float64   `gorm:"type:decimal(12,2)" json:"cash_difference"`
	Status         string     `gorm:"type:varchar(10);not null;index:idx_shifts_user_status" json:"status"` // open, closed
	Notes          string     `gorm:"type:text" json:"notes"`

	User *User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

func (s *Shift) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

// CashMovement mencatat transaksi kas masuk dan kas keluar (petty cash) selama shift berlangsung.
type CashMovement struct {
	ID        string    `gorm:"type:char(36);primaryKey" json:"id"`
	ShiftID   string    `gorm:"type:char(36);not null;index" json:"shift_id"`
	UserID    string    `gorm:"type:char(36);not null;index" json:"user_id"`
	Type      string    `gorm:"type:varchar(10);not null" json:"type"` // cash_in, cash_out
	Amount    float64   `gorm:"type:decimal(12,2);not null" json:"amount"`
	Notes     string    `gorm:"type:varchar(255);not null" json:"notes"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`

	Shift *Shift `gorm:"foreignKey:ShiftID;references:ID" json:"shift,omitempty"`
	User  *User  `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

func (cm *CashMovement) BeforeCreate(tx *gorm.DB) error {
	if cm.ID == "" {
		cm.ID = uuid.New().String()
	}
	return nil
}

// Category merepresentasikan pengelompokan kategori menu.
type Category struct {
	ID       uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	OutletID string  `gorm:"type:char(36);not null;index" json:"outlet_id"`
	Name     string  `gorm:"type:varchar(100);not null" json:"name"`
	Outlet   *Outlet `gorm:"foreignKey:OutletID;references:ID" json:"outlet,omitempty"`
}

// Product merepresentasikan item menu katalog dengan harga jual dan kalkulasi HPP dari BOM.
type Product struct {
	ID          string    `gorm:"type:char(36);primaryKey" json:"id"`
	CategoryID  uint      `gorm:"not null;index" json:"category_id"`
	SKU         string    `gorm:"type:varchar(50)" json:"sku"`
	Name        string    `gorm:"type:varchar(150);not null" json:"name"`
	Price       float64   `gorm:"type:decimal(12,2);not null" json:"price"`
	COGS        float64   `gorm:"type:decimal(12,2);default:0.00" json:"cogs"`
	IsAvailable bool      `gorm:"default:true" json:"is_available"`
	CreatedAt   time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null" json:"updated_at"`

	Category *Category       `gorm:"foreignKey:CategoryID;references:ID" json:"category,omitempty"`
	Recipes  []ProductRecipe `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"recipes,omitempty"`
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// Ingredient merepresentasikan bahan baku mentah yang dipantau untuk kontrol stok BOM.
type Ingredient struct {
	ID            string    `gorm:"type:char(36);primaryKey" json:"id"`
	OutletID      string    `gorm:"type:char(36);not null;index" json:"outlet_id"`
	Name          string    `gorm:"type:varchar(150);not null" json:"name"`
	Unit          string    `gorm:"type:varchar(20);not null" json:"unit"` // gram, ml, pcs, pack
	CurrentStock  float64   `gorm:"type:decimal(10,2);default:0.00" json:"current_stock"`
	MinStockAlert float64   `gorm:"type:decimal(10,2);default:0.00" json:"min_stock_alert"`
	CostPerUnit   float64   `gorm:"type:decimal(12,2);default:0.00" json:"cost_per_unit"`
	UpdatedAt     time.Time `gorm:"not null" json:"updated_at"`

	Outlet *Outlet `gorm:"foreignKey:OutletID;references:ID" json:"outlet,omitempty"`
}

func (i *Ingredient) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = uuid.New().String()
	}
	return nil
}

// ProductRecipe memetakan produk ke bahan baku beserta takaran porsi (Bill of Materials).
type ProductRecipe struct {
	ID               uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID        string  `gorm:"type:char(36);not null;index" json:"product_id"`
	IngredientID     string  `gorm:"type:char(36);not null;index" json:"ingredient_id"`
	QuantityRequired float64 `gorm:"type:decimal(10,2);not null" json:"quantity_required"`

	Product    *Product    `gorm:"foreignKey:ProductID;references:ID;constraint:OnDelete:CASCADE" json:"product,omitempty"`
	Ingredient *Ingredient `gorm:"foreignKey:IngredientID;references:ID;constraint:OnDelete:RESTRICT" json:"ingredient,omitempty"`
}

// InventoryLog mencatat seluruh mutasi stok bahan baku untuk jejak audit dan kartu stok.
type InventoryLog struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	IngredientID   string    `gorm:"type:char(36);not null;index:idx_inventory_logs_ingredient" json:"ingredient_id"`
	OrderItemID    *string   `gorm:"type:char(36);index" json:"order_item_id"`
	ChangeType     string    `gorm:"type:varchar(20);not null" json:"change_type"` // sale_deduction, restock, waste, adjustment
	QuantityChange float64   `gorm:"type:decimal(10,2);not null" json:"quantity_change"`
	RemainingStock float64   `gorm:"type:decimal(10,2);not null" json:"remaining_stock"`
	Notes          string    `gorm:"type:varchar(255)" json:"notes"`
	CreatedAt      time.Time `gorm:"not null;index:idx_inventory_logs_ingredient" json:"created_at"`

	Ingredient *Ingredient `gorm:"foreignKey:IngredientID;references:ID" json:"ingredient,omitempty"`
	OrderItem  *OrderItem  `gorm:"foreignKey:OrderItemID;references:ID" json:"order_item,omitempty"`
}

// Table merepresentasikan meja makan pada restoran.
type Table struct {
	ID          uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	OutletID    string  `gorm:"type:char(36);not null;index" json:"outlet_id"`
	TableNumber string  `gorm:"type:varchar(20);not null" json:"table_number"`
	Capacity    int     `gorm:"default:2" json:"capacity"`
	Status      string  `gorm:"type:varchar(20);default:'available'" json:"status"` // available, occupied, reserved
	Outlet      *Outlet `gorm:"foreignKey:OutletID;references:ID" json:"outlet,omitempty"`
}

// Order merepresentasikan pesanan penjualan utama yang mendukung rekursif pisah tagihan via ParentOrderID.
type Order struct {
	ID             string    `gorm:"type:char(36);primaryKey" json:"id"`
	InvoiceNumber  string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"invoice_number"`
	ParentOrderID  *string   `gorm:"type:char(36);index" json:"parent_order_id"`
	OutletID       string    `gorm:"type:char(36);not null;index" json:"outlet_id"`
	ShiftID        string    `gorm:"type:char(36);not null;index" json:"shift_id"`
	UserID         string    `gorm:"type:char(36);not null;index" json:"user_id"`
	TableID        *uint     `gorm:"index" json:"table_id"`
	ServiceType    string    `gorm:"type:varchar(20);not null" json:"service_type"` // dine_in, take_away, delivery
	CustomerName   string    `gorm:"type:varchar(100)" json:"customer_name"`
	Subtotal       float64   `gorm:"type:decimal(12,2);not null" json:"subtotal"`
	DiscountAmount float64   `gorm:"type:decimal(12,2);default:0.00" json:"discount_amount"`
	TaxAmount      float64   `gorm:"type:decimal(12,2);default:0.00" json:"tax_amount"`
	ServiceAmount  float64   `gorm:"type:decimal(12,2);default:0.00" json:"service_amount"`
	TotalAmount    float64   `gorm:"type:decimal(12,2);not null" json:"total_amount"`
	QueueNumber    int       `gorm:"type:integer;not null;default:0;index:idx_orders_outlet_queue" json:"queue_number"`
	Status         string    `gorm:"type:varchar(20);not null;index:idx_orders_status_created" json:"status"` // open, pending_payment, paid, cancelled, split
	CreatedAt      time.Time `gorm:"not null;index:idx_orders_status_created" json:"created_at"`
	UpdatedAt      time.Time `gorm:"not null" json:"updated_at"`

	Outlet      *Outlet     `gorm:"foreignKey:OutletID;references:ID;constraint:OnDelete:RESTRICT" json:"outlet,omitempty"`
	Shift       *Shift      `gorm:"foreignKey:ShiftID;references:ID" json:"shift,omitempty"`
	User        *User       `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Table       *Table      `gorm:"foreignKey:TableID;references:ID" json:"table,omitempty"`
	ParentOrder *Order      `gorm:"foreignKey:ParentOrderID;references:ID" json:"parent_order,omitempty"`
	OrderItems  []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	Payments    []Payment   `gorm:"foreignKey:OrderID" json:"payments,omitempty"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}

// OrderItem merepresentasikan rincian item dalam pesanan dan melacak status persiapan di KDS dapur.
type OrderItem struct {
	ID            string    `gorm:"type:char(36);primaryKey" json:"id"`
	OrderID       string    `gorm:"type:char(36);not null;index" json:"order_id"`
	ProductID     string    `gorm:"type:char(36);not null;index" json:"product_id"`
	Quantity      int       `gorm:"not null" json:"quantity"`
	UnitPrice     float64   `gorm:"type:decimal(12,2);not null" json:"unit_price"`
	Subtotal      float64   `gorm:"type:decimal(12,2);not null" json:"subtotal"`
	ItemNotes     string    `gorm:"type:varchar(255)" json:"item_notes"`
	KitchenStatus string    `gorm:"type:varchar(20);default:'queued';index:idx_order_items_kitchen" json:"kitchen_status"` // queued, cooking, ready, served
	CreatedAt     time.Time `gorm:"not null;index:idx_order_items_kitchen" json:"created_at"`

	Order   *Order   `gorm:"foreignKey:OrderID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	Product *Product `gorm:"foreignKey:ProductID;references:ID" json:"product,omitempty"`
}

func (oi *OrderItem) BeforeCreate(tx *gorm.DB) error {
	if oi.ID == "" {
		oi.ID = uuid.New().String()
	}
	return nil
}

// Payment mencatat penyelesaian pembayaran tunai dan pembayaran digital gateway Midtrans.
type Payment struct {
	ID                    string     `gorm:"type:char(36);primaryKey" json:"id"`
	OrderID               string     `gorm:"type:char(36);not null;index" json:"order_id"`
	PaymentMethod         string     `gorm:"type:varchar(30);not null" json:"payment_method"` // cash, qris_static, midtrans_dynamic, debit_edc, transfer
	AmountPaid            float64    `gorm:"type:decimal(12,2);not null" json:"amount_paid"`
	ChangeDue             float64    `gorm:"type:decimal(12,2);default:0.00" json:"change_due"`
	MidtransTransactionID string     `gorm:"type:varchar(100)" json:"midtrans_transaction_id"`
	MidtransPaymentType   string     `gorm:"type:varchar(50)" json:"midtrans_payment_type"`
	Status                string     `gorm:"type:varchar(20);not null" json:"status"` // pending, settlement, expired, denied
	SettledAt             *time.Time `json:"settled_at"`
	CreatedAt             time.Time  `gorm:"autoCreateTime" json:"created_at"`

	Order *Order `gorm:"foreignKey:OrderID;references:ID" json:"order,omitempty"`
}

func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
