package graphql

import (
	"context"
	"errors"
	"fmt"
	"nt-folly-xmaxx-comp/internal/app/serve/dataloaders"
	"nt-folly-xmaxx-comp/internal/app/serve/graphql/gqlmodels"

	"github.com/99designs/gqlgen/graphql"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"
)

// Resolver contains the GQL resolvers.
type Resolver struct {
	Conn *pgxpool.Pool
	Log  *zap.Logger
}

/////////////
//  Query  //
/////////////

type queryResolver struct{ *Resolver }

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

// Competition is a query resolver that fetches a specific competition.
func (r *queryResolver) Competition(ctx context.Context, id string) (*gqlmodels.Competition, error) {
	output := &gqlmodels.Competition{}

	// Get competition
	q := `
		SELECT id, name, start_at, finish_at
		FROM competitions
		WHERE id = $1`
	err := r.Conn.QueryRow(ctx, q, id).Scan(&output.ID, &output.Name, &output.StartAt, &output.FinishAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &gqlerror.Error{
			Path:    graphql.GetPath(ctx),
			Message: "Competition not found.",
			Extensions: map[string]interface{}{
				"code": "NOT_FOUND",
			},
		}
	}
	if err != nil {
		return nil, fmt.Errorf("unable to query competition: %w", err)
	}

	// Get involved users
	q = `
		SELECT u.id, u.username, u.display_name, u.membership_type, u.status,
			u.created_at, u.updated_at,
			sum(ur.played) as races,
			round(sum(ur.played) *
				((100.0 + (sum(typed) / 5.0 / (sum(secs) / 60.0)) / 2) *
				(1.0 - sum(ur.errs) / sum(ur.typed::decimal)))) AS points
		FROM users u
			INNER JOIN user_records ur on ur.user_id = u.id 
			INNER JOIN competitions_to_user_records c2ur on c2ur.user_record_id = ur.id AND
				c2ur.competition_id = $1
		GROUP BY u.id
		ORDER BY points DESC`
	rows, err := r.Conn.Query(ctx, q, output.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to query users: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		user := &gqlmodels.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.DisplayName, &user.MembershipType, &user.Status,
			&user.CreatedAt, &user.UpdatedAt,
			&user.Races, &user.Points)
		if err != nil {
			return nil, fmt.Errorf("unable to scan user row: %w", err)
		}
		output.Users = append(output.Users, user)
	}
	return output, nil
}

// RaffleTicketStats is a query resolver that counts number of tickets sold or available.
func (r *queryResolver) RaffleTicketStats(ctx context.Context, id string) (*gqlmodels.RaffleTicketStat, error) {
	if id == "" {
		return nil, &gqlerror.Error{
			Path:    graphql.GetPath(ctx),
			Message: "Raffle ID is required.",
			Extensions: map[string]interface{}{
				"code": "REQUIRED",
			},
		}
	}
	output := &gqlmodels.RaffleTicketStat{}
	q := `
		SELECT count(DISTINCT CASE WHEN rtu.user_id IS NOT NULL THEN rt.code END) AS sold,
			count(DISTINCT CASE WHEN rtu.user_id IS NULL THEN rt.code END) AS available
		from raffle_tickets rt
			LEFT JOIN raffle_ticket_users rtu ON rt.code = rtu.code 
				AND rtu.raffle_id = $1`
	err := r.Conn.QueryRow(ctx, q, id).Scan(&output.Sold, &output.Available)
	if err != nil {
		return nil, fmt.Errorf("querying raffle ticket count failed: %w", err)
	}
	return output, nil
}

// RaffleLogs is a query resolver that fetches all the raffle logs.
func (r *queryResolver) RaffleLogs(ctx context.Context, id string, includeRevoke *bool) ([]*gqlmodels.RaffleTicketLog, error) {
	if id == "" {
		return nil, &gqlerror.Error{
			Path:    graphql.GetPath(ctx),
			Message: "Raffle ID is required.",
			Extensions: map[string]interface{}{
				"code": "REQUIRED",
			},
		}
	}
	q := `
		SELECT rtl.ID, rtl.raffle_id, rtl.user_id, u.username, u.display_name, rtl.code, rtl.action_type, rtl.content, rtl.created_at
		FROM raffle_ticket_logs rtl
			INNER JOIN users u ON u.id = rtl.user_id
		WHERE rtl.raffle_id = $1`
	if includeRevoke != nil && *includeRevoke {
		q += ` AND rtl.action_type != 'REVOKE'`
	}
	rows, err := r.Conn.Query(ctx, q, id)
	if err != nil {
		return nil, fmt.Errorf("query raffle logs failed: %w", err)
	}

	output := []*gqlmodels.RaffleTicketLog{}
	for rows.Next() {
		item := &gqlmodels.RaffleTicketLog{}
		err := rows.Scan(&item.ID, &item.RaffleID, &item.UserID, &item.Username, &item.DisplayName, &item.Code, &item.ActionType, &item.Content, &item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("fetching raffle logs failed: %w", err)
		}
		output = append(output, item)
	}
	return output, nil
}

//////////////////
//  Type: User  //
//////////////////

type userResolver struct{ *Resolver }

func (r *Resolver) User() UserResolver {
	return &userResolver{r}
}

// RaffleTickets is a user query resolver that reveals all of the user's raffle tickets.
func (r *userResolver) RaffleTickets(ctx context.Context, obj *gqlmodels.User, id string) ([]*gqlmodels.RaffleTicketLog, error) {
	dl := dataloaders.GetLoadersFromContext(ctx)
	if dl == nil || dl.RaffleTicketLogsByUserID == nil {
		return nil, fmt.Errorf("unable to find dataloader for user raffle tickets")
	}
	key := dataloaders.IDRaffleUserKey{
		UserID:   obj.ID,
		RaffleID: id,
	}
	output, err := dl.RaffleTicketLogsByUserID.Load(key)
	if err != nil {
		return nil, fmt.Errorf("unable to load user's raffle tickets")
	}
	return output, nil
}
