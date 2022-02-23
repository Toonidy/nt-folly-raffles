package cli

import (
	"fmt"
	"log"
	"nt-folly-xmaxx-comp/internal/pkg/build"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "prints api server build version",
	Long:  "Prints API Server build version.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version:", build.Version)
		fmt.Println("Build Hash:", build.BuildHash)
	},
}

func InitConfig(rootCmd *cobra.Command) func() {
	var cfgFile string

	rootCmd.AddCommand(versionCmd)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.nt-folly-silly-comp)")

	rootCmd.PersistentFlags().Bool("prod", false, "Whether this is used for production.")
	rootCmd.PersistentFlags().String("db_url", "", "database url to connect to (overrides the other db_* values)")
	rootCmd.PersistentFlags().String("db_host", "localhost", "database host")
	rootCmd.PersistentFlags().String("db_port", "5444", "database port")
	rootCmd.PersistentFlags().String("db_user", "nt-folly-comp", "database user")
	rootCmd.PersistentFlags().String("db_pass", "dev", "database password")
	rootCmd.PersistentFlags().String("db_name", "nt-folly-comp", "database name")
	rootCmd.PersistentFlags().String("db_sslmode", "disable", "database connection ssl mode")
	rootCmd.PersistentFlags().Bool("db_debug", false, "whether to log queries for debugging purposes")

	viper.BindPFlag("prod", rootCmd.PersistentFlags().Lookup("prod"))
	viper.BindPFlag("db_url", rootCmd.PersistentFlags().Lookup("db_url"))
	viper.BindPFlag("db_host", rootCmd.PersistentFlags().Lookup("db_host"))
	viper.BindPFlag("db_port", rootCmd.PersistentFlags().Lookup("db_port"))
	viper.BindPFlag("db_user", rootCmd.PersistentFlags().Lookup("db_user"))
	viper.BindPFlag("db_pass", rootCmd.PersistentFlags().Lookup("db_pass"))
	viper.BindPFlag("db_name", rootCmd.PersistentFlags().Lookup("db_name"))
	viper.BindPFlag("db_sslmode", rootCmd.PersistentFlags().Lookup("db_sslmode"))
	viper.BindPFlag("db_debug", rootCmd.PersistentFlags().Lookup("db_sslmode"))

	return func() {
		if cfgFile != "" {
			// Use config file from the flag.
			viper.SetConfigFile(cfgFile)
		} else {
			// Find home directory.
			home, err := homedir.Dir()
			if err != nil {
				log.Fatal(err)
			}

			// Search config in home directory
			viper.AddConfigPath(home)
			viper.SetConfigName(build.DefaultConfigFile)
		}

		viper.AutomaticEnv() // read in environment variables that match

		// If a config file is found, read it in.
		if err := viper.ReadInConfig(); err == nil {
			log.Println("Using config file:", viper.ConfigFileUsed())
		}
	}
}

func CreateLogger() (*zap.Logger, error) {
	var (
		logConfig zap.Config
		err       error
	)
	if viper.GetBool("prod") {
		logConfig = zap.NewProductionConfig()
	} else {
		logConfig = zap.NewDevelopmentConfig()
		logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	logger, err := logConfig.Build()
	if err != nil {
		log.Fatalln("failed to setup logger", err)
	}
	logger = logger.With(zap.String("version", build.Version), zap.String("build", build.BuildHash))
	return logger, nil
}
