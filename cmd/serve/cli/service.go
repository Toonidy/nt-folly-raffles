package cli

import (
	"errors"
	"net/http"
	"nt-folly-xmaxx-comp/internal/app/serve/api"
	"nt-folly-xmaxx-comp/internal/pkg/db"
	"os"
	"os/signal"
	"runtime"
	"strings"

	"github.com/go-chi/cors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// serviceCmd represents the service command.
var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "runs the api server.",
	Long:  "Runs the API Server.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		conn, err := db.ConnectPool(ctx, db.GetConnectionString(), logger)
		if err != nil {
			logger.Error("db connection failed", zap.Error(err))
			return
		}

		// Start API Service
		corsOptions := &cors.Options{
			AllowedOrigins:   strings.Split(viper.GetString("cors_allowed_origins"), ","),
			AllowedMethods:   strings.Split(viper.GetString("cors_allowed_methods"), ","),
			AllowedHeaders:   strings.Split(viper.GetString("cors_allowed_headers"), ","),
			AllowCredentials: viper.GetBool("cors_allow_credentials"),
			MaxAge:           viper.GetInt("cors_max_age"),
		}
		apiAddr := viper.GetString("api_addr")
		apiService := api.NewAPIService(conn, logger, corsOptions)
		server := &http.Server{
			Addr:    apiAddr,
			Handler: apiService,
		}
		go func() {
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Fatal("api - service failed to start", zap.Error(err))
			}
		}()
		logger.Info("api - service started")
		logger.Sugar().Infof("api - hosting on %s", apiAddr)
		logger.Sugar().Infof("api - # cpu: %d", runtime.NumCPU())

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		sig := <-quit
		logger.Info("shutting down server...", zap.Any("reason", sig))

		if err := server.Shutdown(cmd.Context()); err != nil {
			logger.Fatal("api - service failed to shutdown", zap.Error(err))
		}
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)
}
