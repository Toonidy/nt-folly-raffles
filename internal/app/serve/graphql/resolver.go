package graphql

import (
	"context"
	"errors"
	"fmt"
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
