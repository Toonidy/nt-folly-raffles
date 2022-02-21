package cli

import (
	"nt-folly-xmaxx-comp/internal/app/collection/cron"
	"nt-folly-xmaxx-comp/internal/pkg/db"
	"nt-folly-xmaxx-comp/pkg/nitrotype/clients"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Service that collects Nitro Type Team stats.",
	Long:  "Service that collects Nitro Type Team stats. This is done every 10 minutes.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		teamTag := viper.GetString("team_tag")
		if teamTag == "" {
			logger.Error("team_tag is required")
			return
		}
		teamID := viper.GetInt("team_id")
		if teamID <= 0 {
			logger.Error("team_id is required")
			return
		}

		conn, err := db.ConnectPool(ctx, db.GetConnectionString(), logger)
		if err != nil {
			logger.Error("unable to connect to database", zap.Error(err))
			return
		}
		apiClient := clients.NewAPIClientBrowser(viper.GetString("browser_user_agent"))

		// Start Scheduler Service
		logger.Info("cron - service started")
		c := cron.NewCronService(ctx, conn, logger, apiClient, teamTag, teamID)
		c.Start()
		defer c.Stop()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		sig := <-quit
		logger.Info("shutting down scheduler...", zap.Any("reason", sig))
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)
}
