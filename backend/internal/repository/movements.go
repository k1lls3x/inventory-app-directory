package repository

import (
	"context"
	"inventory-app/backend/internal/model"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type MovementRepository interface {
	GetAllMovementsThisMonth(ctx context.Context) ([]model.Movement, error)
}

type PgMovementRepository struct {
	db  *sqlx.DB
	log zerolog.Logger
}

func NewMovementRepository(db *sqlx.DB, log zerolog.Logger) *PgMovementRepository {
	return &PgMovementRepository{
		db:  db,
		log: log,
	}
}
func (r *PgMovementRepository) GetAllMovementsThisMonth(ctx context.Context) ([]model.Movement, error) {
	query := `
		 SELECT
  i.inbound_id      AS movement_id,
  'inbound'         AS type,
  i.item_id         AS item_id,       -- вернуть item_id
  it.name           AS item_name,
  i.quantity,
  i.received_at     AS date,
  i.warehouse_id    AS warehouse_id,  -- вернуть warehouse_id
  w.name            AS warehouse_name,
  s.supplier_id     AS supplier_id,   -- вернуть supplier_id
  s.name            AS supplier_name,
  NULL              AS destination,
  NULL              AS shipped_at     -- чтобы Go не жаловался на пустой столбец
FROM inbound i
JOIN item      it ON i.item_id      = it.item_id
JOIN warehouse w  ON i.warehouse_id = w.warehouse_id
LEFT JOIN supplier s ON i.supplier_id = s.supplier_id
WHERE i.received_at >= date_trunc('month', CURRENT_DATE)

UNION ALL

SELECT
  o.outbound_id     AS movement_id,
  'outbound'        AS type,
  o.item_id         AS item_id,
  it.name           AS item_name,
  o.quantity,
  o.shipped_at      AS date,
  o.warehouse_id    AS warehouse_id,
  w.name            AS warehouse_name,
  NULL              AS supplier_id,
  NULL              AS supplier_name,
  o.destination,
  o.shipped_at      AS shipped_at
FROM outbound o
JOIN item      it ON o.item_id      = it.item_id
JOIN warehouse w  ON o.warehouse_id = w.warehouse_id
WHERE o.shipped_at >= date_trunc('month', CURRENT_DATE)

UNION ALL

SELECT
  t.transfer_id      AS movement_id,
  'transfer'         AS type,
  t.item_id          AS item_id,
  it.name            AS item_name,
  t.quantity,
  t.transferred_at   AS date,
  t.from_warehouse_id AS warehouse_id,
  wf.name || ' → ' || wt.name AS warehouse_name,
  NULL               AS supplier_id,
  NULL               AS supplier_name,
  COALESCE(t.note, '') AS destination,
  NULL               AS shipped_at
FROM stock_transfer t
JOIN item it ON t.item_id = it.item_id
JOIN warehouse wf ON t.from_warehouse_id = wf.warehouse_id
JOIN warehouse wt ON t.to_warehouse_id = wt.warehouse_id
WHERE t.transferred_at >= date_trunc('month', CURRENT_DATE)

ORDER BY date DESC;
	`
	var movements []model.Movement
	err := r.db.SelectContext(ctx, &movements, query)
	if err != nil {
		r.log.Error().Err(err).Msg("Ошибка при получении движений за месяц")
		return nil, err
	}
	r.log.Info().Int("count", len(movements)).Msg("Движения за месяц успешно получены")
	return movements, nil
}
