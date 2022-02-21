package cli

import (
	"database/sql"
	"errors"
	"fmt"
	_ "nt-folly-xmaxx-comp/internal/app/migrate/migrations"
	"nt-folly-xmaxx-comp/internal/pkg/db"
	"strconv"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
)

// dbMigrateUpCmd represents the migration command
var dbMigrateUpCmd = &cobra.Command{
	Use:   "migrate-up",
	Short: "runs db migrate up command.",
	Long:  "Runs DB Migration Up command.",
	Run: func(cmd *cobra.Command, args []string) {
		n := 0
		if len(args) > 0 {
			input, err := strconv.Atoi(args[0])
			if err != nil {
				logger.Error("N arg is invalid", zap.Error(err))
				return
			}
			if input <= 0 {
				logger.Error("N arg is invalid")
				return
			}
			n = input
		}
		conn, err := newDatabaseInstance()
		if err != nil {
			logger.Error("db connection failed", zap.Error(err))
			return
		}
		logger.Info("db migrate up started")
		m, err := newMigrateInstance(conn)
		if err != nil {
			logger.Error("failed to initialize db migration", zap.Error(err))
			return
		}
		if n > 0 {
			err = m.Steps(n)
			if err != nil && !errors.Is(err, migrate.ErrNoChange) {
				logger.Error("failed to db migrate down", zap.Error(err))
				return
			}
		} else {
			err = m.Up()
			if err != nil && !errors.Is(err, migrate.ErrNoChange) {
				logger.Error("failed to db migrate up", zap.Error(err))
				return
			}
		}
		version, dirty, err := m.Version()
		if err != nil {
			logger.Error("failed to db migrate result", zap.Error(err))
			return
		}
		logger.Sugar().Infof("db migrate up finished (version: %d, is dirty: %v)", version, dirty)
	},
}

// dbMigrateDownCmd represents the migration command
var dbMigrateDownCmd = &cobra.Command{
	Use:   "migrate-down", // Setup NT API Client
	Short: "runs DB Migration Down command.",
	Long:  "Runs DB Migration Down command.",
	Run: func(cmd *cobra.Command, args []string) {
		n := 0
		if len(args) > 0 {
			input, err := strconv.Atoi(args[0])
			if err != nil {
				logger.Error("N arg is invalid", zap.Error(err))
				return
			}
			if input <= 0 {
				logger.Error("N arg is invalid")
				return
			}
			n = input
		}
		conn, err := newDatabaseInstance()
		if err != nil {
			logger.Error("db connection failed", zap.Error(err))
			return
		}
		m, err := newMigrateInstance(conn)
		if err != nil {
			logger.Error("failed to initialize db migration", zap.Error(err))
			return
		}
		logger.Info("db migrate down started")
		if n > 0 {
			err = m.Steps(-n)
			if err != nil {
				logger.Error("failed to db migrate down", zap.Error(err))
				return
			}
			version, dirty, err := m.Version()
			if errors.Is(err, migrate.ErrNilVersion) {
				logger.Info("db migrate down finished (version: n/a)")
				return
			}
			if err != nil {
				logger.Error("failed to db migrate result", zap.Error(err))
				return
			}
			logger.Sugar().Infof("db migrate down finished (version: %d, is dirty: %v)", version, dirty)
			return
		}
		err = m.Down()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			logger.Error("failed to db migrate down", zap.Error(err))
			return
		}
		logger.Info("db migrate down finished (version: n/a)")
	},
}

// dbMigrateForceCmd represents the migration command
var dbMigrateForceCmd = &cobra.Command{
	Use:   "migrate-force",
	Short: "runs db migrate force command.",
	Long:  "Runs DB Migration Force command.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			logger.Error("V arg is required")
			return
		}
		v, err := strconv.Atoi(args[0])
		if err != nil {
			logger.Error("N arg is invalid", zap.Error(err))
			return
		}
		if v <= 0 {
			logger.Error("N arg is invalid")
			return
		}
		conn, err := newDatabaseInstance()
		if err != nil {
			logger.Error("db connection failed", zap.Error(err))
			return
		}
		m, err := newMigrateInstance(conn)
		if err != nil {
			logger.Error("failed to initialize db migration", zap.Error(err))
			return
		}
		logger.Info("db migrate force started")
		err = m.Force(v)
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			logger.Error("failed to db migrate force", zap.Error(err))
			return
		}
		version, dirty, err := m.Version()
		if err != nil {
			logger.Error("failed to db migrate result", zap.Error(err))
			return
		}
		logger.Sugar().Infof("db migrate force finished (version: %d, is dirty: %v)", version, dirty)
	},
}

// dbMigrateDropCmd represents the migration command
var dbMigrateDropCmd = &cobra.Command{
	Use:   "migrate-drop",
	Short: "runs db migrate drop command.",
	Long:  "Runs DB Migration Drop command.",
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := newDatabaseInstance()
		if err != nil {
			logger.Error("db connection failed", zap.Error(err))
			return
		}
		m, err := newMigrateInstance(conn)
		if err != nil {
			logger.Error("failed to initialize db migration", zap.Error(err))
			return
		}
		logger.Info("db migrate drop started")
		err = m.Drop()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			logger.Error("failed to db migrate drop", zap.Error(err))
			return
		}
		logger.Info("db migrate drop finished")
	},
}

func init() {
	rootCmd.AddCommand(dbMigrateUpCmd)
	rootCmd.AddCommand(dbMigrateDownCmd)
	rootCmd.AddCommand(dbMigrateForceCmd)
	rootCmd.AddCommand(dbMigrateDropCmd)
}

func newDatabaseInstance() (*sql.DB, error) {
	connString := db.GetConnectionString()
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, fmt.Errorf("unable to open database: %w", err)
	}
	return db, nil
}

func newMigrateInstance(conn *sql.DB) (*migrate.Migrate, error) {
	dbDriver, err := pgx.WithInstance(conn, &pgx.Config{})
	if err != nil {
		return nil, fmt.Errorf("db driver instance: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance("embed://", "pgx", dbDriver)
	if err != nil {
		return nil, fmt.Errorf("migrate instance: %w", err)
	}
	return m, nil
}
