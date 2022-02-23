package cli

import (
	"nt-folly-xmaxx-comp/internal/app/migrate/seed"
	"nt-folly-xmaxx-comp/internal/pkg/db"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// dbSeedCmd represents the migration command
var dbSeedCmd = &cobra.Command{
	Use:   "seed",
	Short: "runs db seed command.",
	Long:  "Runs DB Seed command.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		conn, err := db.ConnectPool(ctx, db.GetConnectionString(), logger)
		if err != nil {
			logger.Error("unable to connect to database", zap.Error(err))
			return
		}
		err = seed.Run(ctx, conn, logger)
		if err != nil {
			logger.Error("db seed failed", zap.Error(err))
			return
		}

		logger.Sugar().Infof("db seed finished")
	},
}

func init() {
	rootCmd.AddCommand(dbSeedCmd)
}
