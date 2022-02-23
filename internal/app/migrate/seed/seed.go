package seed

import (
	"context"
	"fmt"
	"unicode"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

// Run starts the database seeding process.
func Run(ctx context.Context, conn *pgxpool.Pool, logger *zap.Logger) error { // Start Transaction
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}

	// Create raffle tickets
	for letter := 'a'; letter <= 'z'; letter++ {
		for i := 1; i <= 99; i++ {
			code := fmt.Sprintf("%c%d", unicode.ToUpper(letter), i)
			q := `INSERT INTO raffle_tickets (code) VALUES ($1)`
			batch.Queue(q, code)
		}
	}
	br := tx.SendBatch(ctx, batch)
	defer br.Close()
	for i := 0; i < batch.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("insert raffle ticket failed: %w", err)
		}
	}
	br.Close()

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit transaction failed: %w", err)
	}

	return nil
}
