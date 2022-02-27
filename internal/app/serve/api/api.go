package api

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"net/http"
	"nt-folly-xmaxx-comp/internal/app/serve/dataloaders"
	"nt-folly-xmaxx-comp/internal/app/serve/graphql"
	"time"

	gqlgraphql "github.com/99designs/gqlgen/graphql"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"
)

// NewAPIService sets up the API Service for Raffles
func NewAPIService(conn *pgxpool.Pool, log *zap.Logger, corsOptions *cors.Options) http.Handler {
	corsMiddleware := cors.Handler(*corsOptions)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(corsMiddleware)
	r.Use(httprate.LimitByIP(100, 1*time.Minute))
	r.Use(dataloaders.Middleware(conn))

	gqlServer := handler.NewDefaultServer(
		graphql.NewExecutableSchema(
			graphql.Config{
				Resolvers: &graphql.Resolver{
					Conn: conn,
					Log:  log,
				},
			},
		),
	)
	// gqlServer.Use(extension.FixedComplexityLimit(181))
	gqlServer.SetErrorPresenter(
		func(ctx context.Context, e error) *gqlerror.Error {
			if gqlErr, ok := e.(*gqlerror.Error); ok {
				if len(gqlErr.Extensions) > 0 {
					return gqlgraphql.DefaultErrorPresenter(ctx, gqlErr)
				}
			}

			log.Error("graphql internal server error",
				zap.String("reqID", middleware.GetReqID(ctx)),
				zap.Error(e),
			)

			return gqlgraphql.DefaultErrorPresenter(
				ctx,
				gqlerror.Errorf("There was a problem with the server, please try again."),
			)
		},
	)
	gqlServer.SetRecoverFunc(func(ctx context.Context, err interface{}) error {
		log.Error("panic caught",
			zap.String("reqID", middleware.GetReqID(ctx)),
			zap.Any("panic", err),
		)
		return fmt.Errorf("internal server error")
	})
	gqlServer.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	})
	gqlServer.AddTransport(transport.Options{})
	gqlServer.AddTransport(transport.GET{})
	gqlServer.AddTransport(transport.POST{})
	gqlServer.AddTransport(transport.MultipartForm{})
	gqlServer.AroundOperations(func(ctx context.Context, next gqlgraphql.OperationHandler) gqlgraphql.ResponseHandler {
		opCtx := gqlgraphql.GetOperationContext(ctx)
		if opCtx.OperationName != "IntrospectionQuery" && opCtx.Operation != nil {
			rawQueryEncoded := b64.StdEncoding.EncodeToString([]byte(string(opCtx.Operation.Operation) + " " + opCtx.RawQuery))
			log.Info(fmt.Sprintf("graphql request: %s", opCtx.Operation.Operation),
				zap.String("reqID", middleware.GetReqID(ctx)),
				zap.String("gql", rawQueryEncoded),
			)
		}
		return next(ctx)
	})

	r.Route("/api", func(r chi.Router) {
		r.Get("/check", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})
		r.Route("/gql", func(r chi.Router) {
			r.Handle("/", playground.Handler("GraphQL playground", "/api/gql/query"))
			r.Handle("/query", gqlServer)
		})
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Oh Folly"))
	})

	return r
}
