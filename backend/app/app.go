package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"inventory-app/backend/internal/db"
	"inventory-app/backend/internal/handler/export"
	"inventory-app/backend/internal/model"
	"inventory-app/backend/internal/repository"
	"inventory-app/backend/internal/service"
	"os"
	"strings"
	"sync"
	"time"
)

// App struct
type App struct {
	ctx              context.Context
	initOnce         sync.Once
	logger           zerolog.Logger
	itemService      *service.ItemService
	userService      *service.UserService
	authService      *service.AuthService
	inboundService   *service.InboundService
	supplierService  *service.SupplierService
	warehouseService *service.WarehouseService
	stockService     *service.StockService
	outboundService  *service.OutboundService
	dashboardService *service.DashboardService
	movementService  *service.MovementService
}

func NewApp() *App {
	return &App{
		logger: newLogger(),
	}
}

func newLogger() zerolog.Logger {
	logFile, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger.Error().Err(err).Msg("Ошибка открытия файла логов, пишем в stdout")
		return logger
	}

	logger := zerolog.New(logFile).With().Timestamp().Logger()
	logger.Info().Msg("=== Приложение запущено ===")
	return logger
}

func (a *App) initServices() {
	a.initOnce.Do(func() {
		dbInstance := db.Init()

		itemRepo := repository.NewPgItemRepository(dbInstance, a.logger)
		a.itemService = service.NewItemService(itemRepo, a.logger)

		userRepo := repository.NewUserRepository(dbInstance, a.logger)
		a.userService = service.NewUserService(userRepo, a.logger)

		authRepo := repository.NewAuthRepository(dbInstance, a.logger)
		a.authService = service.NewAuthService(authRepo, a.logger)

		inboundRepo := repository.NewInboundRepository(dbInstance, a.logger)
		a.inboundService = service.NewInboundService(inboundRepo, itemRepo, dbInstance, a.logger)

		supplierRepo := repository.NewSupplierRepository(dbInstance, a.logger)
		a.supplierService = service.NewSupplierService(supplierRepo, dbInstance, a.logger)

		warehouseRepo := repository.NewWarehouseRepository(dbInstance, a.logger)
		a.warehouseService = service.NewWarehouseService(warehouseRepo, dbInstance, a.logger)

		stockRepo := repository.NewStockRepository(dbInstance, a.logger)
		a.stockService = service.NewStockService(stockRepo, dbInstance, a.logger)

		outboundRepo := repository.NewOutboundRepository(dbInstance, a.logger)
		a.outboundService = service.NewOutboundService(outboundRepo, dbInstance, a.logger)

		dashboardRepo := repository.NewDashboardRepository(dbInstance, a.logger)
		a.dashboardService = service.NewDashboardService(dashboardRepo, dbInstance, a.logger)

		movementRepo := repository.NewMovementRepository(dbInstance, a.logger)
		a.movementService = service.NewMovementService(movementRepo, dbInstance, a.logger)
	})
}

// Startup is called when the app starts
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.initServices()
}

func (a *App) GetDashboard() (*model.DashboardData, error) {
	return a.dashboardService.LoadDashboard(a.ctx)
}

//-----------------------STOCKS---------------------------------\\

func (a *App) GetStockDetails() ([]model.ItemWithStock, error) {
	return a.stockService.GetStockDetails(a.ctx)
}
func (a *App) GetStocks() ([]model.Stock, error) {
	return a.stockService.GetStocks(a.ctx)
}
func (a *App) AddStock(itemID, quantity, warehouseID int) error {
	return a.stockService.AddStock(a.ctx, itemID, quantity, warehouseID)
}
func (a *App) RemoveStock(stockID int) error {
	return a.stockService.RemoveStock(a.ctx, stockID)
}
func (a *App) FindStockByWarehouse(warehouseID int) ([]model.ItemWithStock, error) {
	return a.stockService.FindStockByWarehouse(a.ctx, warehouseID)
}
func (a *App) ChangeStock(itemID, warehouseID, newQuantity int) error {
	return a.stockService.ChangeStock(a.ctx, itemID, warehouseID, newQuantity)
}
func (a *App) GetWeeklyStockTrend() ([]model.DailyStock, error) {
	return a.stockService.GetWeeklyStockTrend(a.ctx)
}

func (a *App) TransferStock(itemID, fromWarehouseID, toWarehouseID, quantity int, transferredAtStr, note string) error {
	var transferredAt time.Time
	var err error
	if transferredAtStr != "" {
		transferredAt, err = time.Parse("2006-01-02", transferredAtStr)
		if err != nil {
			return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
		}
	}

	var notePtr *string
	if strings.TrimSpace(note) != "" {
		notePtr = &note
	}

	return a.stockService.TransferStock(a.ctx, model.StockTransfer{
		ItemID:          itemID,
		FromWarehouseID: fromWarehouseID,
		ToWarehouseID:   toWarehouseID,
		Quantity:        quantity,
		TransferredAt:   transferredAt,
		Note:            notePtr,
	})
}

func (a *App) GetStockTransfers() ([]model.StockTransferDetails, error) {
	return a.stockService.GetStockTransfers(a.ctx)
}

func (a *App) ExportStockToExcel() (string, error) {
	return Export.ExportStockToExcel(a.stockService, a.logger)
}

func (a *App) ExportUsersToExcel() (string, error) {
	return Export.ExportUsersToExcel(a.userService, a.logger)
}

func (a *App) ExportSuppliersToExcel() (string, error) {
	return Export.ExportSuppliersToExcel(a.supplierService, a.logger)
}

func (a *App) ExportDeliveriesToExcel() (string, error) {
	return Export.ExportDeliveriesToExcel(a.inboundService, a.logger)
}

func (a *App) ExportItemsToExcel() (string, error) {
	return Export.ExportItemsToExcel(a.itemService, a.logger)
}

//-----------------------Items---------------------------------\\

func (a *App) GetItems() ([]model.Item, error) {
	return a.itemService.ListItems(a.ctx)
}

func (a *App) GetAllItems() ([]model.ItemBrief, error) {
	return a.itemService.ListItemsBriefs(a.ctx)
}

func (a *App) GetTopItems() ([]model.ItemWithStock, error) {
	return a.dashboardService.LoadTopItems(a.ctx)
}

func (a *App) UpdateItem(item model.Item) error {
	return a.itemService.UpdateItem(a.ctx, item)
}

func (a *App) RemoveItem(sku string) error {
	return a.itemService.RemoveItem(a.ctx, sku)
}

func (a *App) AddItem(item model.Item) error {
	return a.itemService.AddItem(a.ctx, item)
}

func (a *App) FindItems(filter model.ItemFilter) ([]model.Item, error) {
	return a.itemService.FindItems(a.ctx, filter)
}

// -----------------------Warehouse---------------------------------\\
func (a *App) GetWarehouses() ([]model.Warehouse, error) {
	return a.warehouseService.GetWarehouses(a.ctx)
}
func (a *App) AddWarehouse(warehouse model.Warehouse) error {
	return a.warehouseService.AddWarehouse(a.ctx, warehouse)
}
func (a *App) EditWarehouse(warehouse model.Warehouse) error {
	return a.warehouseService.EditWarehouse(a.ctx, warehouse)
}

func (a *App) GetTurnoverByWarehouse() ([]model.ItemTurnoverByWarehouse, error) {
	return a.dashboardService.LoadTurnoverByWarehouse(a.ctx)
}

//-----------------------Indbound---------------------------------\\

func (a *App) GetInboundDetails() ([]model.InboundDetails, error) {
	return a.inboundService.ListInboundDetails(a.ctx)
}

func (a *App) GetInboundDetailsByDate(date string) ([]model.InboundDetails, error) {
	return a.inboundService.ListInboundDetailsByDate(a.ctx, date)
}

func (a *App) AddInbound(inb model.Inbound) error {
	return a.inboundService.AddInbound(a.ctx, inb)
}
func (a *App) AddInboundTx(inb model.Inbound, item model.Item) error {
	return a.inboundService.AddInboundTx(a.ctx, inb, item)
}
func (a *App) DeleteInbound(inboundId int) error {
	return a.inboundService.DeleteInbound(a.ctx, inboundId)
}

func (a *App) EditInbound(inb model.Inbound) error {
	return a.inboundService.EditInbound(a.ctx, inb)
}

// -----------------------Supplier---------------------------------\\
func (a *App) GetSuppliers() ([]model.Supplier, error) {
	return a.supplierService.GetSuppliers(a.ctx)
}
func (a *App) EditSupplier(sp model.Supplier) error {
	return a.supplierService.EditSupplier(a.ctx, sp)
}
func (a *App) AddSupplier(sp model.Supplier) error {
	return a.supplierService.AddSupplier(a.ctx, sp)
}
func (a *App) RemoveSupplier(supplierId int) error {
	return a.supplierService.RemoveSupplier(a.ctx, supplierId)
}

//-----------------------Auth---------------------------------\\

func (a *App) RegisterUser(username, password, fullName, role string) error {
	user := &model.User{
		Username: username,
		Role:     role,
		FullName: fullName,
	}
	return a.authService.Register(a.ctx, user, password)
}

func (a *App) LoginUser(username, password string) (*model.User, error) {
	user, ok := a.authService.Authorize(a.ctx, username, password)
	if !ok {
		return nil, errors.New("Неверный логин или пароль")
	}
	return user, nil
}

func (a *App) ChangePassword(login, oldPassword, newPassword string) error {
	return a.authService.ChangePassword(a.ctx, login, oldPassword, newPassword)
}

//-----------------------Users---------------------------------\\

func (a *App) RemoveUser(userId int) error {
	return a.userService.RemoveUser(a.ctx, userId)
}

func (a *App) GetUsers() ([]model.User, error) {
	return a.userService.GetUsers(a.ctx)
}

func (a *App) ChangeUserData(u model.UserUpdate) error {
	return a.userService.ChangeUserData(a.ctx, u)
}

// -----------------------Movements---------------------------------\\
func (a *App) GetAllMovementsThisMonth() ([]model.Movement, error) {
	return a.movementService.GetAllMovementsThisMonth(a.ctx)
}

//-----------------------Outbound---------------------------------\\

func (a *App) GetOutboundDetails() ([]model.OutboundDetails, error) {
	return a.outboundService.GetOutboundDetails(a.ctx)
}

func (a *App) AddOutbound(itemID int, quantity int, shippedAtStr string, destination string, warehouseID int, shippedBy int, documentNo string, note string) error {
	var shippedAt time.Time
	var err error
	if shippedAtStr != "" {
		shippedAt, err = time.Parse("2006-01-02", shippedAtStr)
		if err != nil {
			return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
		}
	}
	var shippedByPtr *int
	if shippedBy > 0 {
		shippedByPtr = &shippedBy
	}
	var documentNoPtr *string
	if strings.TrimSpace(documentNo) != "" {
		documentNoPtr = &documentNo
	}
	var notePtr *string
	if strings.TrimSpace(note) != "" {
		notePtr = &note
	}
	return a.outboundService.AddOutbound(a.ctx, model.Outbound{
		ItemID:      itemID,
		Quantity:    quantity,
		ShippedAt:   shippedAt,
		ShippedBy:   shippedByPtr,
		Destination: &destination,
		DocumentNo:  documentNoPtr,
		Note:        notePtr,
		WarehouseID: warehouseID,
	})
}

func (a *App) EditOutbound(
	itemID int,
	quantity int,
	shippedAtStr string,
	destination string,
	warehouseID int,
	outboundID int,
	shippedBy int,
	documentNo string,
	note string,
) error {
	var shippedAt time.Time
	var err error
	if shippedAtStr != "" {
		shippedAt, err = time.Parse("2006-01-02", shippedAtStr)
		if err != nil {
			return fmt.Errorf("invalid date %q, expected YYYY-MM-DD: %w", shippedAtStr, err)
		}
	}
	var shippedByPtr *int
	if shippedBy > 0 {
		shippedByPtr = &shippedBy
	}
	var documentNoPtr *string
	if strings.TrimSpace(documentNo) != "" {
		documentNoPtr = &documentNo
	}
	var notePtr *string
	if strings.TrimSpace(note) != "" {
		notePtr = &note
	}
	return a.outboundService.EditOutbound(a.ctx, model.Outbound{
		OutboundID:  outboundID,
		ItemID:      itemID,
		Quantity:    quantity,
		ShippedAt:   shippedAt,
		ShippedBy:   shippedByPtr,
		Destination: &destination,
		DocumentNo:  documentNoPtr,
		Note:        notePtr,
		WarehouseID: warehouseID,
	})
}

func (a *App) RemoveOutbound(outboundId int) error {
	return a.outboundService.DeleteOutbound(a.ctx, outboundId)
}
