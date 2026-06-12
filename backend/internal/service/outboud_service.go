package service

import (
	"context"
	"fmt"
	"inventory-app/backend/internal/model"
	"inventory-app/backend/internal/repository"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type OutboundService struct {
	repo repository.OutboundRepository
	log  zerolog.Logger
	db   *sqlx.DB
}

func NewOutboundService(repo repository.OutboundRepository, db *sqlx.DB, log zerolog.Logger) *OutboundService {
	return &OutboundService{
		repo: repo,
		log:  log,
		db:   db,
	}
}

func (s *OutboundService) GetOutboundDetails(ctx context.Context) ([]model.OutboundDetails, error) {
	items, err := s.repo.GetOutboundDetails(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("Ошибка при получении отгрузок")
		return nil, err
	}
	s.log.Info().Int("count", len(items)).Msg("Список отгрузок успешно получен")
	return items, nil
}

func (s *OutboundService) AddOutbound(ctx context.Context, outb model.Outbound) error {
	if outb.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := adjustStockTx(ctx, tx, outb.ItemID, outb.WarehouseID, -outb.Quantity); err != nil {
		return fmt.Errorf("reserve outbound stock: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO outbound (item_id, quantity, shipped_at, destination, warehouse_id, shipped_by, document_no, note)
		VALUES ($1, $2, COALESCE($3, NOW()), $4, $5, $6, $7, $8)
	`, outb.ItemID, outb.Quantity, nullableTime(outb.ShippedAt), outb.Destination, outb.WarehouseID, outb.ShippedBy, outb.DocumentNo, outb.Note)
	if err != nil {
		s.log.Error().Err(err).Msg("Ошибка при добавлении отгрузки")
		return fmt.Errorf("add outbound: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit outbound: %w", err)
	}

	s.log.Info().
		Int("item_id", outb.ItemID).
		Int("warehouse_id", outb.WarehouseID).
		Msg("Отгрузка успешно добавлена")
	return nil
}

func (s *OutboundService) GetOutboundDetailsByDate(ctx context.Context, date string) ([]model.OutboundDetails, error) {
	items, err := s.repo.GetOutboundDetailsByDate(ctx, date)
	if err != nil {
		s.log.Error().Str("date", date).Err(err).Msg("Ошибка при получении отгрузок по дате")
		return nil, err
	}
	s.log.Info().Str("date", date).Int("count", len(items)).Msg("Список отгрузок по дате успешно получен")
	return items, nil
}

func (s *OutboundService) DeleteOutbound(ctx context.Context, outboundId int) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var old model.Outbound
	err = tx.GetContext(ctx, &old, `
		SELECT outbound_id, item_id, quantity, shipped_at, destination, warehouse_id
		FROM outbound
		WHERE outbound_id = $1
		FOR UPDATE
	`, outboundId)
	if err != nil {
		return fmt.Errorf("load outbound: %w", err)
	}

	if err := adjustStockTx(ctx, tx, old.ItemID, old.WarehouseID, old.Quantity); err != nil {
		return fmt.Errorf("reverse outbound stock: %w", err)
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM outbound WHERE outbound_id = $1`, outboundId)
	if err != nil {
		s.log.Error().Int("outbound_id", outboundId).Err(err).Msg("Ошибка при удалении отгрузки")
		return fmt.Errorf("delete outbound: %w", err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("check deleted outbound: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("outbound %d not found", outboundId)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete outbound: %w", err)
	}
	s.log.Info().Int("outbound_id", outboundId).Msg("Отгрузка успешно удалена")
	return nil
}

func (s *OutboundService) EditOutbound(ctx context.Context, outb model.Outbound) error {
	if outb.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var old model.Outbound
	err = tx.GetContext(ctx, &old, `
		SELECT outbound_id, item_id, quantity, shipped_at, destination, warehouse_id
		FROM outbound
		WHERE outbound_id = $1
		FOR UPDATE
	`, outb.OutboundID)
	if err != nil {
		return fmt.Errorf("load outbound: %w", err)
	}

	if old.ItemID == outb.ItemID && old.WarehouseID == outb.WarehouseID {
		if err := adjustStockTx(ctx, tx, outb.ItemID, outb.WarehouseID, old.Quantity-outb.Quantity); err != nil {
			return fmt.Errorf("apply outbound stock delta: %w", err)
		}
	} else {
		if err := adjustStockTx(ctx, tx, old.ItemID, old.WarehouseID, old.Quantity); err != nil {
			return fmt.Errorf("reverse old outbound stock: %w", err)
		}
		if err := adjustStockTx(ctx, tx, outb.ItemID, outb.WarehouseID, -outb.Quantity); err != nil {
			return fmt.Errorf("apply new outbound stock: %w", err)
		}
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE outbound
		SET item_id = $1,
			warehouse_id = $2,
			quantity = $3,
			shipped_at = COALESCE($4, shipped_at),
			destination = $5,
			shipped_by = $6,
			document_no = $7,
			note = $8
		WHERE outbound_id = $9
	`, outb.ItemID, outb.WarehouseID, outb.Quantity, nullableTime(outb.ShippedAt), outb.Destination, outb.ShippedBy, outb.DocumentNo, outb.Note, outb.OutboundID)
	if err != nil {
		s.log.Error().Int("outbound_id", outb.OutboundID).Err(err).Msg("Ошибка при изменении данных отгрузки")
		return fmt.Errorf("update outbound: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit edit outbound: %w", err)
	}
	s.log.Info().Int("outbound_id", outb.OutboundID).Msg("Данные отгрузки успешно изменены")
	return nil
}
