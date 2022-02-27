package dataloaders

import (
	"context"
	"fmt"
	"net/http"
	"nt-folly-xmaxx-comp/internal/app/serve/graphql/gqlmodels"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type contextKey string

const key = contextKey("dataloaders")

// Loaders hold references to the individual dataloaders.
type Loaders struct {
	RaffleTicketLogsByUserID *RaffleTicketLogLoader
}

// newLoaders initializes individual loaders.
func newLoaders(ctx context.Context, conn *pgxpool.Pool) *Loaders {
	return &Loaders{
		RaffleTicketLogsByUserID: raffleTicketLogsLoader(conn),
	}
}

// GetLoadersFromContext retrives dataloaders from context
func GetLoadersFromContext(ctx context.Context) *Loaders {
	return ctx.Value(key).(*Loaders)
}

// Middleware stores Loaders as a requested-scored context value.
func Middleware(conn *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			loaders := newLoaders(ctx, conn)
			augmentedCtx := context.WithValue(ctx, key, loaders)
			r = r.WithContext(augmentedCtx)
			next.ServeHTTP(w, r)
		})
	}
}

///////////////////
//  Dataloaders  //
///////////////////

// raffleTicketLogsLoader fetches the raffle ticket logs for the following resolver:
// * user -> raffleTickets
func raffleTicketLogsLoader(conn *pgxpool.Pool) *RaffleTicketLogLoader {
	return NewRaffleTicketLogLoader(
		RaffleTicketLogLoaderConfig{
			Fetch: func(ids []IDRaffleUserKey) ([][]*gqlmodels.RaffleTicketLog, []error) {
				if len(ids) == 0 {
					return [][]*gqlmodels.RaffleTicketLog{}, nil
				}

				output := [][]*gqlmodels.RaffleTicketLog{}
				args := []interface{}{}
				raffleTicketIDCondition := []string{}
				userIDCondition := []string{}

				for _, key := range ids {
					output = append(output, []*gqlmodels.RaffleTicketLog{})
					raffleTicketIDCondition = append(raffleTicketIDCondition, fmt.Sprintf("$%d", len(args)+1))
					userIDCondition = append(userIDCondition, fmt.Sprintf("$%d", len(args)+2))
					args = append(args, key.RaffleID, key.UserID)
				}

				q := fmt.Sprintf(`
					SELECT rtl.ID, rtu.raffle_id, rtu.user_id, rtu.code, rtl.action_type, rtl.content
					FROM raffle_ticket_users rtu 
						INNER JOIN raffle_ticket_logs rtl ON rtl.raffle_id = rtu.raffle_id 
							AND rtl.user_id = rtu.user_id 
							AND rtl.code = rtu.code 
							AND rtl.action_type = 'GIVE'
					WHERE rtu.raffle_id IN (%s)
						AND rtu.user_id IN (%s)`,
					strings.Join(raffleTicketIDCondition, ","),
					strings.Join(userIDCondition, ","),
				)
				rows, err := conn.Query(context.Background(), q, args...)
				if err != nil {
					return nil, []error{fmt.Errorf("unable to query user raffle tickets: %w", err)}
				}
				defer rows.Close()

				for rows.Next() {
					item := &gqlmodels.RaffleTicketLog{}
					err := rows.Scan(&item.ID, &item.RaffleID, &item.UserID, &item.Code, &item.ActionType, &item.Content)
					if err != nil {
						return nil, []error{fmt.Errorf("unable to read user raffle ticket record: %w", err)}
					}
					for i, key := range ids {
						if key.RaffleID == item.RaffleID && key.UserID == item.UserID {
							output[i] = append(output[i], item)
							break
						}
					}
				}

				return output, nil
			},
			Wait:     1 * time.Millisecond,
			MaxBatch: 100,
		},
	)
}
