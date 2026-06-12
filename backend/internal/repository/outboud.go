package repository

import (
	"context"
	"inventory-app/backend/internal/model"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type OutboundRepository interface {
	GetOutboundDetails(ctx context.Context) ([]model.OutboundDetails, error)
	AddOutbound(ctx context.Context, outb model.Outbound) error
	GetOutboundDetailsByDate(ctx context.Context, date string) ([]model.OutboundDetails, error)
	DeleteOutbound(ctx context.Context, outboundId int) error
	EditOutbound(ctx context.Context, outb model.Outbound) error
}

type PgOutboundRepository struct {
	db  *sqlx.DB
	log zerolog.Logger
}

func NewOutboundRepository(db *sqlx.DB, log zerolog.Logger) *PgOutboundRepository {
	return &PgOutboundRepository{db: db, log: log}
}

// Получить все отгрузки с деталями
func (r *PgOutboundRepository) GetOutboundDetails(ctx context.Context) ([]model.OutboundDetails, error) {
	var items []model.OutboundDetails
	query := `
		SELECT
			o.outbound_id,
			o.item_id,
			o.warehouse_id,
			o.shipped_by,
			to_char(o.shipped_at, 'YYYY-MM-DD') AS date,
			to_char(o.shipped_at, 'YYYY-MM-DD') AS shipped_at,
			it.name AS name,
			it.sku AS sku,
			o.destination AS destination,
			o.quantity AS quantity,
			w.name AS warehouse,
			COALESCE(NULLIF(u.full_name, ''), u.username, '') AS shipper_name,
			COALESCE(o.document_no, '') AS document_no,
			COALESCE(o.note, '') AS note
		FROM outbound o
		JOIN item it       ON o.item_id = it.item_id
		JOIN warehouse w   ON o.warehouse_id = w.warehouse_id
		LEFT JOIN "User" u ON o.shipped_by = u.user_id
		ORDER BY o.shipped_at DESC;
	`
	err := r.db.SelectContext(ctx, &items, query)
	if err != nil {
		r.log.Error().Err(err).Msg("Ошибка при получении отгрузок")
		return nil, err
	}
	r.log.Info().Int("count", len(items)).Msg("Список отгрузок успешно получен")
	return items, nil
}

// Добавить отгрузку
func (r *PgOutboundRepository) AddOutbound(ctx context.Context, outb model.Outbound) error {
	query := `
		INSERT INTO outbound (item_id, quantity, shipped_at, destination, warehouse_id, shipped_by, document_no, note)
		VALUES ($1, $2, COALESCE($3, now()), $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query, outb.ItemID, outb.Quantity, outb.ShippedAt, outb.Destination, outb.WarehouseID, outb.ShippedBy, outb.DocumentNo, outb.Note)
	if err != nil {
		r.log.Error().
			Int("item_id", outb.ItemID).
			Int("warehouse_id", outb.WarehouseID).
			Err(err).Msg("Ошибка при добавлении отгрузки")
		return err
	}
	r.log.Info().
		Int("item_id", outb.ItemID).
		Int("warehouse_id", outb.WarehouseID).
		Msg("Отгрузка успешно добавлена")
	return nil
}

// Получить отгрузки по дате
func (r *PgOutboundRepository) GetOutboundDetailsByDate(ctx context.Context, date string) ([]model.OutboundDetails, error) {
	var items []model.OutboundDetails
	query := `
		SELECT
			o.outbound_id,
			o.item_id,
			o.warehouse_id,
			o.shipped_by,
			to_char(o.shipped_at, 'YYYY-MM-DD') AS date,
			to_char(o.shipped_at, 'YYYY-MM-DD') AS shipped_at,
			it.name AS name,
			it.sku AS sku,
			o.destination AS destination,
			o.quantity AS quantity,
			w.name AS warehouse,
			COALESCE(NULLIF(u.full_name, ''), u.username, '') AS shipper_name,
			COALESCE(o.document_no, '') AS document_no,
			COALESCE(o.note, '') AS note
		FROM outbound o
		JOIN item it ON o.item_id = it.item_id
		JOIN warehouse w ON o.warehouse_id = w.warehouse_id
		LEFT JOIN "User" u ON o.shipped_by = u.user_id
		WHERE DATE(o.shipped_at) = $1
		ORDER BY o.shipped_at DESC;
	`
	err := r.db.SelectContext(ctx, &items, query, date)
	if err != nil {
		r.log.Error().Str("date", date).Err(err).Msg("Ошибка при получении отгрузок по дате")
		return nil, err
	}
	r.log.Info().Str("date", date).Int("count", len(items)).Msg("Список отгрузок по дате успешно получен")
	return items, nil
}

// Удалить отгрузку
func (r *PgOutboundRepository) DeleteOutbound(ctx context.Context, outboundId int) error {
	query := `DELETE FROM outbound WHERE outbound_id = $1;`
	res, err := r.db.ExecContext(ctx, query, outboundId)
	if err != nil {
		r.log.Error().Int("outbound_id", outboundId).Err(err).Msg("Ошибка при удалении отгрузки")
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		r.log.Warn().Int("outbound_id", outboundId).Msg("Отгрузка не найдена для удаления")
	} else {
		r.log.Info().Int("outbound_id", outboundId).Msg("Отгрузка успешно удалена")
	}
	return nil
}

// Редактировать отгрузку
func (r *PgOutboundRepository) EditOutbound(ctx context.Context, outb model.Outbound) error {
	query := `
		UPDATE outbound
		SET item_id = $1,
			warehouse_id = $2,
			quantity = $3,
			shipped_at = $4,
			destination = $5,
			shipped_by = $6,
			document_no = $7,
			note = $8
		WHERE outbound_id = $9;
	`
	_, err := r.db.ExecContext(ctx, query, outb.ItemID, outb.WarehouseID, outb.Quantity, outb.ShippedAt, outb.Destination, outb.ShippedBy, outb.DocumentNo, outb.Note, outb.OutboundID)
	if err != nil {
		r.log.Error().Int("outbound_id", outb.OutboundID).Err(err).Msg("Ошибка при изменении данных отгрузки")
		return err
	}
	r.log.Info().Int("outbound_id", outb.OutboundID).Msg("Данные отгрузки успешно изменены")
	return nil
}
