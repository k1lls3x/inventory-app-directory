package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

func nullableTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

func adjustStockTx(ctx context.Context, tx *sqlx.Tx, itemID, warehouseID, delta int) error {
	if itemID <= 0 {
		return fmt.Errorf("item_id must be positive")
	}
	if warehouseID <= 0 {
		return fmt.Errorf("warehouse_id must be positive")
	}
	if delta == 0 {
		return nil
	}

	if delta > 0 {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO stock (item_id, warehouse_id, quantity, last_updated)
			VALUES ($1, $2, $3, NOW())
			ON CONFLICT (item_id, warehouse_id)
			DO UPDATE SET
				quantity = stock.quantity + EXCLUDED.quantity,
				last_updated = NOW()
		`, itemID, warehouseID, delta)
		if err != nil {
			return fmt.Errorf("increase stock: %w", err)
		}
		return nil
	}

	required := -delta
	var current int
	err := tx.GetContext(ctx, &current, `
		SELECT quantity
		FROM stock
		WHERE item_id = $1 AND warehouse_id = $2
		FOR UPDATE
	`, itemID, warehouseID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("остаток по товару %d на складе %d не найден", itemID, warehouseID)
	}
	if err != nil {
		return fmt.Errorf("load stock: %w", err)
	}
	if current < required {
		return fmt.Errorf("недостаточно остатка: доступно %d, требуется %d", current, required)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE stock
		SET quantity = quantity + $3,
			last_updated = NOW()
		WHERE item_id = $1 AND warehouse_id = $2
	`, itemID, warehouseID, delta)
	if err != nil {
		return fmt.Errorf("decrease stock: %w", err)
	}

	return nil
}
