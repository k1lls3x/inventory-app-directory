package service

import (
	"context"
	"fmt"
	"inventory-app/backend/internal/model"
	"inventory-app/backend/internal/repository"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type StockService struct {
	repo repository.StockRepository
	log  zerolog.Logger
	db   *sqlx.DB
}

func NewStockService(repo repository.StockRepository, db *sqlx.DB, log zerolog.Logger) *StockService {
	return &StockService{
		repo: repo,
		log:  log,
		db:   db,
	}
}

func (s *StockService) GetStockDetails(ctx context.Context) ([]model.ItemWithStock, error) {
	result, err := s.repo.GetStockDetails(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("Ошибка при получении данных остатков")
		return nil, err
	}
	s.log.Info().Int("count", len(result)).Msg("Остатки успешно получены")
	return result, nil
}

func (s *StockService) GetWeeklyStockTrend(ctx context.Context) ([]model.DailyStock, error) {
	trend, err := s.repo.GetWeeklyStockTrend(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("Ошибка при получении weekly stock trend")
		return nil, err
	}
	s.log.Info().Int("count", len(trend)).Msg("Weekly stock trend успешно получен")
	return trend, nil
}

func (s *StockService) AddStock(ctx context.Context, itemID, quantity, warehouseID int) error {
	if quantity < 0 {
		return fmt.Errorf("quantity must not be negative")
	}

	err := s.repo.AddStock(ctx, itemID, quantity, warehouseID)
	if err != nil {
		s.log.Error().
			Int("item_id", itemID).
			Int("warehouse_id", warehouseID).
			Err(err).Msg("Ошибка при добавлении/обновлении остатка")
		return err
	}
	s.log.Info().
		Int("item_id", itemID).
		Int("warehouse_id", warehouseID).
		Msg("Остаток успешно добавлен/обновлён")
	return nil
}

func (s *StockService) FindStockByWarehouse(ctx context.Context, warehouseID int) ([]model.ItemWithStock, error) {
	result, err := s.repo.FindStockByWarehouse(ctx, warehouseID)
	if err != nil {
		s.log.Error().Int("warehouse_id", warehouseID).Err(err).Msg("Ошибка при получении остатков по складу")
		return nil, err
	}
	s.log.Info().Int("warehouse_id", warehouseID).Int("count", len(result)).Msg("Остатки по складу успешно получены")
	return result, nil
}

func (s *StockService) ChangeStock(ctx context.Context, itemID, warehouseID, newQuantity int) error {
	if newQuantity < 0 {
		return fmt.Errorf("quantity must not be negative")
	}

	err := s.repo.ChangeStock(ctx, itemID, warehouseID, newQuantity)
	if err != nil {
		s.log.Error().
			Int("item_id", itemID).
			Int("warehouse_id", warehouseID).
			Int("new_quantity", newQuantity).
			Err(err).Msg("Ошибка при изменении stock")
		return err
	}
	s.log.Info().
		Int("item_id", itemID).
		Int("warehouse_id", warehouseID).
		Int("new_quantity", newQuantity).
		Msg("Остаток успешно изменён")
	return nil
}

func (s *StockService) RemoveStock(ctx context.Context, stockID int) error {
	err := s.repo.RemoveStock(ctx, stockID)
	if err != nil {
		s.log.Error().Int("stock_id", stockID).Err(err).Msg("Ошибка при удалении stock")
		return err
	}
	s.log.Info().Int("stock_id", stockID).Msg("Запись stock успешно удалена")
	return nil
}

func (s *StockService) GetStocks(ctx context.Context) ([]model.Stock, error) {
	result, err := s.repo.GetStocks(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("Ошибка при получении всех stock")
		return nil, err
	}
	s.log.Info().Int("count", len(result)).Msg("Все остатки успешно получены")
	return result, nil
}

func (s *StockService) TransferStock(ctx context.Context, transfer model.StockTransfer) error {
	if transfer.ItemID <= 0 {
		return fmt.Errorf("item_id must be positive")
	}
	if transfer.FromWarehouseID <= 0 || transfer.ToWarehouseID <= 0 {
		return fmt.Errorf("warehouse_id must be positive")
	}
	if transfer.FromWarehouseID == transfer.ToWarehouseID {
		return fmt.Errorf("нельзя переместить товар на тот же склад")
	}
	if transfer.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := adjustStockTx(ctx, tx, transfer.ItemID, transfer.FromWarehouseID, -transfer.Quantity); err != nil {
		return fmt.Errorf("decrease source stock: %w", err)
	}
	if err := adjustStockTx(ctx, tx, transfer.ItemID, transfer.ToWarehouseID, transfer.Quantity); err != nil {
		return fmt.Errorf("increase destination stock: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO stock_transfer (
			item_id,
			from_warehouse_id,
			to_warehouse_id,
			quantity,
			transferred_at,
			note
		) VALUES ($1, $2, $3, $4, COALESCE($5, NOW()), $6)
	`, transfer.ItemID, transfer.FromWarehouseID, transfer.ToWarehouseID, transfer.Quantity, nullableTime(transfer.TransferredAt), transfer.Note)
	if err != nil {
		return fmt.Errorf("insert stock transfer: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit stock transfer: %w", err)
	}

	s.log.Info().
		Int("item_id", transfer.ItemID).
		Int("from_warehouse_id", transfer.FromWarehouseID).
		Int("to_warehouse_id", transfer.ToWarehouseID).
		Int("quantity", transfer.Quantity).
		Msg("Межскладское перемещение успешно создано")
	return nil
}

func (s *StockService) GetStockTransfers(ctx context.Context) ([]model.StockTransferDetails, error) {
	var result []model.StockTransferDetails
	err := s.db.SelectContext(ctx, &result, `
		SELECT
			t.transfer_id,
			t.item_id,
			t.from_warehouse_id,
			t.to_warehouse_id,
			to_char(t.transferred_at, 'YYYY-MM-DD') AS date,
			i.name AS item_name,
			i.sku,
			wf.name AS from_warehouse,
			wt.name AS to_warehouse,
			t.quantity,
			COALESCE(t.note, '') AS note
		FROM stock_transfer t
		JOIN item i ON i.item_id = t.item_id
		JOIN warehouse wf ON wf.warehouse_id = t.from_warehouse_id
		JOIN warehouse wt ON wt.warehouse_id = t.to_warehouse_id
		ORDER BY t.transferred_at DESC, t.transfer_id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("load stock transfers: %w", err)
	}

	return result, nil
}
