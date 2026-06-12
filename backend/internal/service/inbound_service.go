package service

import (
	"context"
	"fmt"
	"inventory-app/backend/internal/model"
	"inventory-app/backend/internal/repository"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type InboundService struct {
	repo     repository.InboundRepository
	itemRepo repository.ItemRepository
	log      zerolog.Logger
	db       *sqlx.DB
}

func NewInboundService(repo repository.InboundRepository, itemRepo repository.ItemRepository, db *sqlx.DB, log zerolog.Logger) *InboundService {
	return &InboundService{
		repo:     repo,
		itemRepo: itemRepo,
		log:      log,
		db:       db,
	}
}

func (s *InboundService) AddInboundTx(ctx context.Context, inb model.Inbound, item model.Item) error {
	if inb.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		s.log.Error().Err(err).Msg("Ошибка открытия транзакции")
		return err
	}
	defer tx.Rollback()

	if inb.ItemID <= 0 {
		itemExists, err := s.itemRepo.ItemExistsTx(ctx, tx, item.SKU)
		if err != nil {
			return fmt.Errorf("check item exists: %w", err)
		}
		if !itemExists {
			if err := s.itemRepo.AddItemTx(ctx, tx, item); err != nil {
				return fmt.Errorf("add item: %w", err)
			}
		}
		if err := tx.GetContext(ctx, &inb.ItemID, `SELECT item_id FROM item WHERE sku = $1`, item.SKU); err != nil {
			return fmt.Errorf("load item id: %w", err)
		}
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO inbound (item_id, supplier_id, quantity, received_at, warehouse_id, received_by, document_no, note)
		VALUES ($1, $2, $3, COALESCE($4, NOW()), $5, $6, $7, $8)
	`, inb.ItemID, inb.SupplierID, inb.Quantity, nullableTime(inb.ReceivedAt), inb.WarehouseID, inb.ReceivedBy, inb.DocumentNo, inb.Note)
	if err != nil {
		return fmt.Errorf("add inbound: %w", err)
	}

	if err := adjustStockTx(ctx, tx, inb.ItemID, inb.WarehouseID, inb.Quantity); err != nil {
		return fmt.Errorf("update stock after inbound: %w", err)
	}

	if err := tx.Commit(); err != nil {
		s.log.Error().Err(err).Msg("Ошибка коммита транзакции")
		return err
	}
	return nil
}

func (s *InboundService) ListInboundDetails(ctx context.Context) ([]model.InboundDetails, error) {
	inbounds, err := s.repo.GetInboundDetails(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("Ошибка при получении списка поставок")
		return nil, err
	}
	s.log.Info().Int("count", len(inbounds)).Msg("Список поставок успешно получен")
	return inbounds, nil
}

func (s *InboundService) AddInbound(ctx context.Context, inb model.Inbound) error {
	if inb.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO inbound (item_id, supplier_id, quantity, received_at, warehouse_id, received_by, document_no, note)
		VALUES ($1, $2, $3, COALESCE($4, NOW()), $5, $6, $7, $8)
	`, inb.ItemID, inb.SupplierID, inb.Quantity, nullableTime(inb.ReceivedAt), inb.WarehouseID, inb.ReceivedBy, inb.DocumentNo, inb.Note)
	if err != nil {
		s.log.Error().
			Int("item_id", inb.ItemID).
			Int("supplier_id", inb.SupplierID).
			Int("quantity", inb.Quantity).
			Str("received_at", inb.ReceivedAt.Format("2006-01-02")).
			Int("warehouse_id", inb.WarehouseID).
			Err(err).
			Msg("Ошибка при добавлении поставки (AddInboundSimple)")
		return fmt.Errorf("add inbound: %w", err)
	}
	if err := adjustStockTx(ctx, tx, inb.ItemID, inb.WarehouseID, inb.Quantity); err != nil {
		return fmt.Errorf("update stock after inbound: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit inbound: %w", err)
	}

	s.log.Info().
		Int("item_id", inb.ItemID).
		Int("supplier_id", inb.SupplierID).
		Int("quantity", inb.Quantity).
		Str("received_at", inb.ReceivedAt.Format("2006-01-02")).
		Int("warehouse_id", inb.WarehouseID).
		Msg("Поставка успешно добавлена (AddInboundSimple)")
	return nil
}

func (s *InboundService) ListInboundDetailsByDate(ctx context.Context, date string) ([]model.InboundDetails, error) {
	inbounds, err := s.repo.GetInboundDetailsByDate(ctx, date)
	if err != nil {
		s.log.Error().Err(err).Str("date", date).Msg("Ошибка при получении поставок по дате")
		return nil, err
	}
	s.log.Info().Str("date", date).Int("count", len(inbounds)).Msg("Список поставок по дате успешно получен")
	return inbounds, nil
}

func (s *InboundService) DeleteInbound(ctx context.Context, inboundId int) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var old model.Inbound
	err = tx.GetContext(ctx, &old, `
		SELECT inbound_id, item_id, supplier_id, quantity, received_at, warehouse_id
		FROM inbound
		WHERE inbound_id = $1
		FOR UPDATE
	`, inboundId)
	if err != nil {
		return fmt.Errorf("load inbound: %w", err)
	}

	if err := adjustStockTx(ctx, tx, old.ItemID, old.WarehouseID, -old.Quantity); err != nil {
		return fmt.Errorf("reverse inbound stock: %w", err)
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM inbound WHERE inbound_id = $1`, inboundId)
	if err != nil {
		s.log.Error().
			Int("inbound_id", inboundId).
			Err(err).
			Msg("Ошибка при удалении поставки")
		return fmt.Errorf("delete inbound: %w", err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("check deleted inbound: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("inbound %d not found", inboundId)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete inbound: %w", err)
	}
	s.log.Info().
		Int("inbound_id", inboundId).
		Msg("Поставка успешно удалена")
	return nil
}

func (s *InboundService) EditInbound(ctx context.Context, inb model.Inbound) error {
	if inb.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var old model.Inbound
	err = tx.GetContext(ctx, &old, `
		SELECT inbound_id, item_id, supplier_id, quantity, received_at, warehouse_id
		FROM inbound
		WHERE inbound_id = $1
		FOR UPDATE
	`, inb.InboundID)
	if err != nil {
		s.log.Error().
			Int("inbound_id", inb.InboundID).
			Err(err).
			Msg("Ошибка при редактировании поставки")
		return fmt.Errorf("load inbound: %w", err)
	}

	if old.ItemID == inb.ItemID && old.WarehouseID == inb.WarehouseID {
		if err := adjustStockTx(ctx, tx, inb.ItemID, inb.WarehouseID, inb.Quantity-old.Quantity); err != nil {
			return fmt.Errorf("apply inbound stock delta: %w", err)
		}
	} else {
		if err := adjustStockTx(ctx, tx, old.ItemID, old.WarehouseID, -old.Quantity); err != nil {
			return fmt.Errorf("reverse old inbound stock: %w", err)
		}
		if err := adjustStockTx(ctx, tx, inb.ItemID, inb.WarehouseID, inb.Quantity); err != nil {
			return fmt.Errorf("apply new inbound stock: %w", err)
		}
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE inbound
		SET item_id = $1,
			supplier_id = $2,
			warehouse_id = $3,
			quantity = $4,
			received_at = COALESCE($5, received_at),
			received_by = $6,
			document_no = $7,
			note = $8
		WHERE inbound_id = $9
	`, inb.ItemID, inb.SupplierID, inb.WarehouseID, inb.Quantity, nullableTime(inb.ReceivedAt), inb.ReceivedBy, inb.DocumentNo, inb.Note, inb.InboundID)
	if err != nil {
		return fmt.Errorf("update inbound: %w", err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("check updated inbound: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("inbound %d not found", inb.InboundID)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit edit inbound: %w", err)
	}
	s.log.Info().
		Int("inbound_id", inb.InboundID).
		Msg("Поставка успешно отредактирована")
	return nil
}
