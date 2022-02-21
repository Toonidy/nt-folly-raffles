package cli

import (
	"fmt"
	"log"
	"nt-folly-xmaxx-comp/internal/pkg/cli"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var logger *zap.Logger

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nt-folly-xmaxx-comp-serve",
	Short: "Nitro Type Folly Team Xmaxx Comp Data API Server.",
	Long:  "Nitro Type Folly Team Xmaxx Comp Data API Server.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Define Default Configuration
	rootCmd.PersistentFlags().String("api_addr", ":8080", "api server to host as")
	rootCmd.PersistentFlags().String("cors_allowed_origins", "*", "allowed origins to access CORS")
	rootCmd.PersistentFlags().String("cors_allowed_methods", "GET,POST,PUT,DELETE,OPTIONS", "allowed http methods for CORS")
	rootCmd.PersistentFlags().String("cors_allowed_headers", "Accept,Authorization,Cache-Control,Content-Type,DNT,If-Modified-Since,Keep-Alive,Origin,User-Agent,X-Requested-With", "allowed http headers for CORS")
	rootCmd.PersistentFlags().Bool("cors_allow_credentials", true, "whether to allow credentials for CORS")
	rootCmd.PersistentFlags().Int("cors_max_age", 1728000, "TTL to cache CORS")

	viper.BindPFlag("api_addr", rootCmd.PersistentFlags().Lookup("api_addr"))
	viper.BindPFlag("cors_allowed_origins", rootCmd.PersistentFlags().Lookup("cors_allowed_origins"))
	viper.BindPFlag("cors_allowed_methods", rootCmd.PersistentFlags().Lookup("cors_allowed_methods"))
	viper.BindPFlag("cors_allowed_headers", rootCmd.PersistentFlags().Lookup("cors_allowed_headers"))
	viper.BindPFlag("cors_allow_credentials", rootCmd.PersistentFlags().Lookup("cors_allow_credentials"))
	viper.BindPFlag("cors_max_age", rootCmd.PersistentFlags().Lookup("cors_max_age"))

	// Setup CLI
	cobra.OnInitialize(cli.InitConfig(rootCmd), func() {
		var err error
		logger, err = cli.CreateLogger()
		if err != nil {
			log.Fatalln("unable to setup logger")
		}
	})
}
