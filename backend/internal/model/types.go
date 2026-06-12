package model

import "time"

type User struct {
	UserID       int    `db:"user_id" json:"user_id"`
	Username     string `db:"username" json:"username"`
	PasswordHash string `db:"password_hash" json:"-"`
	FullName     string `db:"full_name" json:"full_name"`
	Role         string `db:"role" json:"role"`
}

type Item struct {
	ItemID       int             `db:"item_id" json:"item_id"`
	SKU          string          `db:"sku" json:"sku"`
	Name         string          `db:"name" json:"name"`
	Description  *string         `db:"description" json:"description"`
	UOM          string          `db:"uom" json:"uom"`
	ReorderLevel int             `db:"reorder_level" json:"reorder_level"`
	ReorderQty   int             `db:"reorder_qty" json:"reorder_qty"`
	Price        JSONNullFloat64 `db:"price" json:"price"`
	Cost         JSONNullFloat64 `db:"cost" json:"cost"`
	CreatedAt    time.Time       `db:"created_at" json:"created_at"`
}

type Supplier struct {
	SupplierID    int     `db:"supplier_id" json:"supplier_id"`
	Name          string  `db:"name" json:"name"`
	Inn           *string `db:"inn" json:"inn"`
	ContactPerson *string `db:"contact_person" json:"contact_person"`
	Phone         *string `db:"phone" json:"phone"`
	Email         *string `db:"email" json:"email"`
}

type Inbound struct {
	InboundID   int       `db:"inbound_id" json:"inbound_id"`
	ItemID      int       `db:"item_id" json:"item_id"`
	SupplierID  int       `db:"supplier_id" json:"supplier_id"`
	Quantity    int       `db:"quantity" json:"quantity"`
	ReceivedAt  time.Time `db:"received_at" json:"received_at"`
	ReceivedBy  *int      `db:"received_by" json:"received_by"`
	DocumentNo  *string   `db:"document_no" json:"document_no"`
	Note        *string   `db:"note" json:"note"`
	WarehouseID int       `db:"warehouse_id" json:"warehouse_id"`
}

type Outbound struct {
	OutboundID  int       `db:"outbound_id" json:"outbound_id"`
	ItemID      int       `db:"item_id" json:"item_id"`
	Quantity    int       `db:"quantity" json:"quantity"`
	ShippedAt   time.Time `db:"shipped_at" json:"shipped_at"`
	ShippedBy   *int      `db:"shipped_by" json:"shipped_by"`
	Destination *string   `db:"destination" json:"destination"`
	DocumentNo  *string   `db:"document_no" json:"document_no"`
	Note        *string   `db:"note" json:"note"`
	WarehouseID int       `db:"warehouse_id" json:"warehouse_id"`
}

type Stock struct {
	StockID     int       `db:"stock_id" json:"stock_id"`
	ItemID      int       `db:"item_id" json:"item_id"`
	Quantity    int       `db:"quantity" json:"quantity"`
	WarehouseID int       `db:"warehouse_id" json:"warehouse_id"`
	LastUpdated time.Time `db:"last_updated" json:"last_updated"`
}

type StockTransfer struct {
	TransferID      int       `db:"transfer_id" json:"transfer_id"`
	ItemID          int       `db:"item_id" json:"item_id"`
	FromWarehouseID int       `db:"from_warehouse_id" json:"from_warehouse_id"`
	ToWarehouseID   int       `db:"to_warehouse_id" json:"to_warehouse_id"`
	Quantity        int       `db:"quantity" json:"quantity"`
	TransferredAt   time.Time `db:"transferred_at" json:"transferred_at"`
	Note            *string   `db:"note" json:"note"`
}

type StockTransferDetails struct {
	TransferID      int    `db:"transfer_id" json:"transfer_id"`
	ItemID          int    `db:"item_id" json:"item_id"`
	FromWarehouseID int    `db:"from_warehouse_id" json:"from_warehouse_id"`
	ToWarehouseID   int    `db:"to_warehouse_id" json:"to_warehouse_id"`
	Date            string `db:"date" json:"date"`
	ItemName        string `db:"item_name" json:"item_name"`
	SKU             string `db:"sku" json:"sku"`
	FromWarehouse   string `db:"from_warehouse" json:"from_warehouse"`
	ToWarehouse     string `db:"to_warehouse" json:"to_warehouse"`
	Quantity        int    `db:"quantity" json:"quantity"`
	Note            string `db:"note" json:"note"`
}

type Warehouse struct {
	WarehouseID int     `db:"warehouse_id" json:"warehouse_id"`
	Name        string  `db:"name" json:"name"`
	Location    *string `db:"location" json:"location"`
}

// ---------------------------------------------------------------\\
type ItemTurnoverByWarehouse struct {
	Warehouse string `db:"warehouse" json:"warehouse"`
	Name      string `db:"name" json:"name"`
	SKU       string `db:"sku" json:"sku"`
	Received  int    `db:"received" json:"received"`
	Shipped   int    `db:"shipped" json:"shipped"`
}

type Movement struct {
	Type          string     `db:"type"            json:"type"`
	MovementID    int        `db:"movement_id"     json:"movement_id"`
	ItemID        int        `db:"item_id"         json:"item_id"`
	ItemName      string     `db:"item_name"       json:"item_name"`
	Quantity      int        `db:"quantity"        json:"quantity"`
	Date          time.Time  `db:"date"            json:"date"`
	WarehouseID   int        `db:"warehouse_id"    json:"warehouse_id"`
	WarehouseName string     `db:"warehouse_name"  json:"warehouse_name"`
	SupplierID    *int       `db:"supplier_id"     json:"supplier_id,omitempty"`
	SupplierName  *string    `db:"supplier_name"   json:"supplier_name,omitempty"`
	Destination   *string    `db:"destination"     json:"destination,omitempty"`
	ShippedAt     *time.Time `db:"shipped_at"      json:"shipped_at,omitempty"`
}

type ItemWithStock struct {
	StockID     int    `json:"stock_id" db:"stock_id"`
	ItemID      int    `json:"item_id" db:"item_id"`
	WarehouseID int    `json:"warehouse_id" db:"warehouse_id"`
	Name        string `json:"name" db:"name"`
	SKU         string `json:"sku" db:"sku"`
	Warehouse   string `json:"warehouse" db:"warehouse"`
	Quantity    int    `json:"quantity" db:"quantity"`
}

type ItemFilter struct {
	SKU      *string
	Name     *string
	UOM      *string
	PriceMin *float64
	PriceMax *float64
}

type ItemBrief struct {
	ItemID int    `db:"item_id" json:"item_id"`
	Name   string `db:"name" json:"name"`
	SKU    string `db:"sku" json:"sku"`
}

type DailyStock struct {
	Date  string `json:"date"`
	Total int    `json:"total"`
}

type InboundDetails struct {
	InboundID    int    `db:"inbound_id" json:"inbound_id"`
	ItemID       int    `db:"item_id" json:"item_id"`
	SupplierID   int    `db:"supplier_id" json:"supplier_id"`
	WarehouseID  int    `db:"warehouse_id" json:"warehouse_id"`
	ReceivedBy   *int   `db:"received_by" json:"received_by"`
	Date         string `db:"date" json:"date"`
	ReceivedAt   string `db:"received_at" json:"received_at"`
	Name         string `db:"name" json:"name"`
	Sku          string `db:"sku" json:"sku"`
	Supplier     string `db:"supplier" json:"supplier"`
	Quantity     int    `db:"quantity" json:"quantity"`
	Warehouse    string `db:"warehouse" json:"warehouse"`
	ReceiverName string `db:"receiver_name" json:"receiver_name"`
	DocumentNo   string `db:"document_no" json:"document_no"`
	Note         string `db:"note" json:"note"`
}

type DashboardData struct {
	TotalStock    float64 `json:"total_stock"`
	ItemCount     int     `json:"item_count"`
	MonthlyOrders int     `json:"monthly_orders"`
	NewItems      int     `json:"new_items"`
}

type UserUpdate struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
	Password string `json:"password,omitempty"` // ← фронт присылает только если хочет сменить пароль
}

type OutboundDetails struct {
	OutboundID  int    `db:"outbound_id" json:"outbound_id"`
	ItemID      int    `db:"item_id" json:"item_id"`
	WarehouseID int    `db:"warehouse_id" json:"warehouse_id"`
	ShippedBy   *int   `db:"shipped_by" json:"shipped_by"`
	Date        string `db:"date" json:"date"`
	ShippedAt   string `db:"shipped_at" json:"shipped_at"`
	Name        string `db:"name" json:"name"`
	Sku         string `db:"sku" json:"sku"`
	Destination string `db:"destination" json:"destination"`
	Quantity    int    `db:"quantity" json:"quantity"`
	Warehouse   string `db:"warehouse" json:"warehouse"`
	ShipperName string `db:"shipper_name" json:"shipper_name"`
	DocumentNo  string `db:"document_no" json:"document_no"`
	Note        string `db:"note" json:"note"`
}
