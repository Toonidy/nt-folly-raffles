package cli

import (
	"fmt"
	"log"
	"nt-folly-xmaxx-comp/internal/pkg/cli"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var logger *zap.Logger

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nt-folly-xmaxx-comp-migrate",
	Short: "Database migration tool for nt-folly-xmaxx-comp.",
	Long:  "Database migration tool for nt-folly-xmaxx-comp.",
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
	cobra.OnInitialize(cli.InitConfig(rootCmd), func() {
		var err error
		logger, err = cli.CreateLogger()
		if err != nil {
			log.Fatalln("unable to setup logger")
		}
	})
}
