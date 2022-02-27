package seed

import (
	"context"
	"fmt"
	"time"
	"unicode"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

// Run starts the database seeding process.
func Run(ctx context.Context, conn *pgxpool.Pool, logger *zap.Logger) error { // Start Transaction
	// Begin Transaction
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

	// Competitions
	loc, err := time.LoadLocation("Australia/Perth")
	if err != nil {
		return fmt.Errorf("timezone load failed: %w", err)
	}
	timeFrom := time.Date(2022, 2, 23, 21, 31, 0, 0, loc)
	timeTo := time.Date(2022, 3, 4, 21, 31, 0, 0, loc)

	var raffleID string
	q := `INSERT INTO raffles (name, start_at, finish_at) VALUES ($1, $2, $3) RETURNING id`
	err = tx.QueryRow(ctx, q, "Main Event", timeFrom, timeTo).Scan(&raffleID)
	if err != nil {
		return fmt.Errorf("insert main event failed: %w", err)
	}

	q = `INSERT INTO competitions (name, start_at, finish_at) VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, q, "Main Event", timeFrom, timeTo)
	if err != nil {
		return fmt.Errorf("insert main event failed: %w", err)
	}

	// Raffle Daily Bonuses
	dailyFrom := timeFrom
	dailyTo := timeFrom.Add(time.Hour * 24)
	for i := 0; i < 7; i++ {
		var competitionID string
		q := `INSERT INTO competitions (name, start_at, finish_at) VALUES ($1, $2, $3) RETURNING id`
		err := tx.QueryRow(ctx, q, fmt.Sprintf("Raffle Bonus Day %d", i+1), dailyFrom, dailyTo).Scan(&competitionID)
		if err != nil {
			return fmt.Errorf("insert bonus competition event failed: %w", err)
		}
		dailyFrom = dailyFrom.Add(time.Hour * 24)
		dailyTo = dailyTo.Add(time.Hour * 24)

		q = `INSERT INTO raffle_bonus_competitions (raffle_id, competition_id) VALUES ($1, $2)`
		_, err = tx.Exec(ctx, q, raffleID, competitionID)
		if err != nil {
			return fmt.Errorf("linking bonus competition to raffles failed: %w", err)
		}
	}

	// Commit Transaction
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit transaction failed: %w", err)
	}

	return nil
}
