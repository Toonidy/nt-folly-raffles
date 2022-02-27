package cron

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"nt-folly-xmaxx-comp/internal/pkg/utils"
	"nt-folly-xmaxx-comp/pkg/nitrotype"
	"time"

	"github.com/go-logr/zapr"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

const (
	TicketPrice      = 5
	ActionTypeGive   = "GIVE"
	ActionTypeRevoke = "REVOKE"
)

type RaffleUser struct {
	ID       string
	RaffleID string
	Amount   int
}

// NewCronService creates a new cron service ready to be activated
func NewCronService(ctx context.Context, conn *pgxpool.Pool, log *zap.Logger, apiClient nitrotype.APIClient, teamTag string, teamID int) *cron.Cron {
	logger := zapr.NewLogger(log)
	c := cron.New(
		cron.WithChain(cron.DelayIfStillRunning(logger)),
	)
	c.AddFunc("1,11,21,31,41,51 * * * *", syncTeams(ctx, conn, log, apiClient, teamTag, teamID))
	return c
}

// syncTeams is the scheduled task function that collect Nitro Type Team Logs.
func syncTeams(ctx context.Context, conn *pgxpool.Pool, log *zap.Logger, apiClient nitrotype.APIClient, teamTag string, teamID int) func() {
	log = log.With(
		zap.String("job", "syncTeams"),
		zap.String("team", teamTag),
	)

	return func() {
		now := time.Now()
		now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())
		defer func() {
			if r := recover(); r != nil {
				log.Error("recovering from panic", zap.Any("panic", r))
			}
		}()

		log.Info("sync teams started")

		// Get Previous Log
		var (
			prevRequestID pgtype.UUID
			prevLogID     pgtype.UUID
		)
		q := `
			SELECT id, api_team_log_id
			FROM nt_api_team_log_requests
			WHERE deleted_at IS NULL
				AND response_type IN ('NEW', 'CACHE')
			ORDER BY created_at DESC
			LIMIT 1`
		err := conn.QueryRow(ctx, q).Scan(&prevRequestID, &prevLogID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			log.Error("unable to query previous log", zap.Error(err))
			return
		}

		// Grab Latest Stats
		teamData, err := apiClient.GetTeam(teamTag)
		if err != nil || teamData.Status != "OK" || teamData.Results.Info == nil {
			log.Error("unable to pull team log", zap.Error(err))

			// Record Fail Request
			if prevLogID.Status == pgtype.Present && prevRequestID.Status == pgtype.Present {
				responseType := "ERROR"
				description := "Unknown error"
				if err != nil {
					description = err.Error()
				} else if teamData.Status != "OK" || teamData.Results.Info == nil {
					description = "Team API Request Failed"
				}
				var newLogID pgtype.UUID
				q = `
					INSERT INTO nt_api_team_log_requests (prev_id, api_team_log_id, response_type, description)
					VALUES ($1, $2, $3, $4)
					RETURNING id`
				err = conn.QueryRow(ctx, q, prevRequestID, prevLogID, responseType, description).Scan(&newLogID)
				if err != nil {
					log.Error("unable to insert request log (error)", zap.Error(err))
				}
			}
			return
		}

		// Check if data doesn't matches team
		if teamID != teamData.Results.Info.TeamID {
			log.Error("team has changed", zap.Int("teamID", teamData.Results.Info.TeamID))
			return
		}

		// Calculate Hash
		data, err := json.Marshal(teamData)
		if err != nil {
			log.Error("unable to marshal team data", zap.Error(err))
			return
		}
		hash, err := utils.HashData(data)
		if err != nil {
			log.Error("unable to calculate team data hash", zap.Error(err))
			return
		}

		// Insert Team Log
		tx, err := conn.Begin(ctx)
		if err != nil {
			log.Error("unable to start recording team data", zap.Error(err))
			return
		}
		defer tx.Rollback(ctx)

		logID := ""
		responseType := "NEW"
		description := "New log download"
		q = `SELECT id FROM nt_api_team_logs WHERE hash = $1`
		err = tx.QueryRow(ctx, q, hash).Scan(&logID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			log.Error("unable to find existing team log", zap.Error(err))
			return
		}
		if logID == "" {
			q := `
				INSERT INTO nt_api_team_logs (hash, log_data)
				VALUES ($1, $2)
				ON CONFLICT (hash) DO NOTHING
				RETURNING id`
			err = tx.QueryRow(ctx, q, hash, data).Scan(&logID)
			if err != nil {
				log.Error("unable to insert team log", zap.Error(err))
				return
			}
		}
		if logID == "" {
			log.Error("unable to find team log id (blank data)")
			return
		}

		if prevLogID.Status == pgtype.Present {
			prevLogIDText := ""
			prevLogID.AssignTo(&prevLogIDText)
			if prevLogIDText == logID {
				responseType = "CACHE"
				description = "Same log found"
			}
		}
		if prevRequestID.Status != pgtype.Present {
			prevRequestID.Set(nil)
		}

		// Insert Team Log Request
		var newLogID string
		q = `
			INSERT INTO nt_api_team_log_requests (prev_id, api_team_log_id, response_type, description)
			VALUES ($1, $2, $3, $4)
			RETURNING id`
		err = tx.QueryRow(ctx, q, prevRequestID, logID, responseType, description).Scan(&newLogID)
		if err != nil {
			log.Error("unable to insert team log request", zap.Error(err))
			return
		}

		// Commit Transaction
		err = tx.Commit(ctx)
		if err != nil {
			log.Error("unable to finish recording team data", zap.Error(err))
			return
		}

		// Calculate Stats (if there was a previous record)
		if prevRequestID.Status == pgtype.Present && newLogID != "" {
			tx, err := conn.Begin(ctx)
			if err != nil {
				log.Error("unable to start team member stats ", zap.Error(err))
				return
			}
			defer tx.Rollback(ctx)

			// Record or Update members
			q = `
				INSERT INTO users (reference_id, username, display_name, membership_type, status)

				SELECT (m->>'userID')::int AS reference_id,
					m->>'username' AS username,
					(
						CASE 
							WHEN m->>'displayName' IS NOT NULL AND m->>'displayName' != '' THEN m->>'displayName'
							ELSE m->>'username'
						END
					) AS display_name,
					(
						CASE m->>'membership'
							WHEN 'gold' THEN 'GOLD'
							ELSE 'BASIC'
						END
					) AS membership_type,
					'NEW' AS status
				FROM nt_api_team_log_requests r
					INNER JOIN nt_api_team_logs l ON l.id = r.api_team_log_id AND json_typeof(l.log_data->'results'->'members') = 'array'
					INNER JOIN json_array_elements(l.log_data->'results'->'members') AS m ON m->>'userID' IS NOT NULL
				WHERE r.id = $1
				
				ON CONFLICT (reference_id) DO UPDATE
				SET username = EXCLUDED.username,
					display_name = EXCLUDED.display_name,
					membership_type = EXCLUDED.membership_type,
					status = CASE WHEN users.status = 'LEFT' THEN EXCLUDED.status ELSE users.status END`
			_, err = tx.Exec(ctx, q, newLogID)
			if err != nil {
				log.Error("unable to update team member details", zap.Error(err))
				return
			}

			// Insert in the records
			q = `
				INSERT INTO user_records (request_id, user_id, played, typed, errs, secs, from_at, to_at)
				SELECT $1 AS request_id,
					(
						SELECT _u.id
						FROM users _u
						WHERE _u.reference_id = (m1->>'userID')::int
						LIMIT 1
					) AS user_id,
					((m1->>'played')::int - (m2->>'played')::int) AS played,
					((m1->>'typed')::int - (m2->>'typed')::int) AS typed,
					((m1->>'errs')::int - (m2->>'errs')::int) AS errs,
					((m1->>'secs')::int - (m2->>'secs')::int) AS secs,
					r2.created_at AS from_at,
					r1.created_at AS to_at
				FROM nt_api_team_log_requests r1				
					INNER JOIN nt_api_team_log_requests r2 ON r2.id = r1.prev_id
						AND r2.api_team_log_id != r1.api_team_log_id
					INNER JOIN nt_api_team_logs l1 ON l1.id = r1.api_team_log_id AND json_typeof(l1.log_data->'results'->'members') = 'array'
					INNER JOIN nt_api_team_logs l2 ON l2.id = r2.api_team_log_id AND json_typeof(l2.log_data->'results'->'members') = 'array'
					INNER JOIN json_array_elements(l1.log_data->'results'->'members') AS m1 ON m1->>'userID' IS NOT NULL
					INNER JOIN json_array_elements(l2.log_data->'results'->'members') AS m2 ON (m1->>'userID')::int = (m2->>'userID')::int
				WHERE r1.id = $1
					AND r1.prev_id IS NOT NULL
					AND ((m1->>'played')::int - (m2->>'played')::int) > 0`
			_, err = tx.Exec(ctx, q, newLogID)
			if err != nil {
				log.Error("unable to insert team member records", zap.Error(err))
				return
			}

			// Update user participation status
			q = `
				UPDATE users u
				SET status = 'ACTIVE', updated_at = NOW()
				WHERE status = 'NEW'
					AND EXISTS (
						SELECT 1
						FROM user_records _r 
						WHERE _r.request_id = $1 
							AND _r.user_id = u.id
						LIMIT 1
					)`
			_, err = tx.Exec(ctx, q, newLogID)
			if err != nil {
				log.Error("unable to update team member update status", zap.Error(err))
				return
			}

			// Update users disqualified status
			q = `
				UPDATE users u
				SET status = 'LEFT',
					updated_at = NOW()
				FROM (
					SELECT (m2->>'userID')::int AS reference_id
					FROM nt_api_team_log_requests r1				
						INNER JOIN nt_api_team_log_requests r2 ON r2.id = r1.prev_id
							AND r2.api_team_log_id != r1.api_team_log_id
						INNER JOIN nt_api_team_logs l1 ON l1.id = r1.api_team_log_id AND json_typeof(l1.log_data->'results'->'members') = 'array'
						INNER JOIN nt_api_team_logs l2 ON l2.id = r2.api_team_log_id AND json_typeof(l2.log_data->'results'->'members') = 'array'
						LEFT JOIN json_array_elements(l2.log_data->'results'->'members') AS m2 ON m2->>'userID' IS NOT NULL
						LEFT JOIN json_array_elements(l1.log_data->'results'->'members') AS m1 ON (m1->>'userID')::int = (m2->>'userID')::int
					WHERE r1.id = $1
						AND r1.prev_id IS NOT NULL
						AND (
							m1->>'userID' IS NULL
							OR m1->>'status' = 'banned'
						)
				) l
				WHERE u.status != 'LEFT'
					AND u.reference_id = l.reference_id`
			_, err = tx.Exec(ctx, q, newLogID)
			if err != nil {
				log.Error("unable to update team member update status", zap.Error(err))
				return
			}

			// Map user records to competitions
			q = `
				INSERT INTO competitions_to_user_records
				
				SELECT c.id AS competition_id, ur.id AS user_record_id
				FROM competitions c, user_records ur 
				WHERE ur.request_id = $1
					AND c.deleted_at IS NULL
					AND c.start_at + INTERVAL '10 minute' <= $2
					AND c.finish_at + INTERVAL '1 minute' >= $2
				
				ON CONFLICT (competition_id, user_record_id) DO NOTHING`
			_, err = tx.Exec(ctx, q, newLogID, now)
			if err != nil {
				log.Error("unable to update team member update status", zap.Error(err))
				return
			}

			// Check Racing Progress for raffle ticket credits...
			q = `
				INSERT INTO raffle_users (raffle_id, user_id, balance)

				SELECT r.id AS raffle_id, ur.user_id, ur.played AS balance
				FROM raffles r, user_records ur
				WHERE ur.request_id = $1
					AND r.deleted_at IS NULL
					AND r.start_at + INTERVAL '10 minute' <= $2
					AND r.finish_at + INTERVAL '1 minute' >= $2
				
				ON CONFLICT (raffle_id, user_id) DO UPDATE
				SET balance = raffle_users.balance + EXCLUDED.balance`
			_, err = tx.Exec(ctx, q, newLogID, now)
			if err != nil {
				log.Error("unable to update team member update status", zap.Error(err))
				return
			}

			// Issue Raffle Tickets
			processRaffleUsers := []RaffleUser{}
			q = `SELECT raffle_id, user_id, balance FROM raffle_users WHERE balance >= $1`
			rows, err := tx.Query(ctx, q, TicketPrice)
			if err != nil {
				log.Error("unable to find raffle ticket users", zap.Error(err))
				return
			}
			defer rows.Close()
			for rows.Next() {
				user := RaffleUser{}
				err := rows.Scan(&user.RaffleID, &user.ID, &user.Amount)
				if err != nil {
					log.Error("unable to check raffle ticket user's balance ", zap.Error(err))
					return
				}
				user.Amount /= TicketPrice

				processRaffleUsers = append(processRaffleUsers, user)

			}

			for _, user := range processRaffleUsers {
				q := fmt.Sprintf(`
					INSERT INTO raffle_ticket_users (raffle_id, user_id, code)

					SELECT $1 AS raffle_id, $2 AS user_id, rt.code
					FROM raffle_tickets rt
						LEFT JOIN raffle_ticket_users rtu ON rtu.raffle_id = $1
							AND rtu.code = rt.code
					WHERE rtu.user_id IS NULL
					ORDER BY rt.sort_index ASC
					LIMIT %d
					RETURNING code`,
					user.Amount,
				)
				ticketCodes := []string{}
				rows, err = tx.Query(ctx, q, user.RaffleID, user.ID)
				if err != nil {
					log.Error("unable to give raffle tickets", zap.Error(err))
					return
				}
				for rows.Next() {
					var code string
					err := rows.Scan(&code)
					if err != nil {
						log.Error("unable to track given raffle ticket", zap.Error(err))
						return
					}
					ticketCodes = append(ticketCodes, code)
				}
				rows.Close()

				for _, code := range ticketCodes {
					q = `
						INSERT INTO raffle_ticket_logs (raffle_id, user_id, code, action_type, content)
						VALUES ($1, $2, $3, $4, $5)`
					_, err = tx.Exec(ctx, q, user.RaffleID, user.ID, code, ActionTypeGive, fmt.Sprintf("Completed %d races.", TicketPrice))
					if err != nil {
						log.Error("unable to log raffle tickets", zap.Error(err))
						return
					}
				}
			}

			q = `
				UPDATE raffle_users
				SET balance = balance - ($1 * (balance / $1))
				WHERE balance >= $1`
			_, err = tx.Exec(ctx, q, TicketPrice)
			if err != nil {
				log.Error("unable to update raffle user balances", zap.Error(err))
				return
			}

			// Free up raffle tickets after disqualifying players
			q = `
				INSERT INTO raffle_ticket_logs (raffle_id, user_id, code, action_type, content)
				SELECT rtu.raffle_id, rtu.user_id, rtu.code, $1 AS action_type, 'Player Left team' AS content
				FROM raffle_ticket_users rtu
					INNER JOIN users u ON u.id = rtu.user_id
				WHERE u.status = 'LEFT'`
			_, err = tx.Exec(ctx, q, ActionTypeRevoke)
			if err != nil {
				log.Error("unable to log upcoming revoked raffle tickets", zap.Error(err))
				return
			}
			q = `
				DELETE
				FROM raffle_ticket_users
				WHERE user_id IN (
					SELECT id
					FROM users
					WHERE status = 'LEFT'
				)`
			_, err = tx.Exec(ctx, q)
			if err != nil {
				log.Error("unable to update raffle user balances", zap.Error(err))
				return
			}
			q = `
				DELETE
				FROM raffle_users
				WHERE user_id IN (
					SELECT id
					FROM users
					WHERE status = 'LEFT'
				)`
			_, err = tx.Exec(ctx, q)
			if err != nil {
				log.Error("unable to update raffle user balances", zap.Error(err))
				return
			}

			// Commit Transaction
			err = tx.Commit(ctx)
			if err != nil {
				log.Error("unable to finish team member stats ", zap.Error(err))
				return
			}
		}

		log.Info("sync teams completed")
	}
}
