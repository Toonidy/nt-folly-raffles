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
	Use:   "nt-folly-xmaxx-comp-collection",
	Short: "Nitro Type Folly Team Xmaxx Comp Data Collection.",
	Long:  "Nitro Type Folly Team Xmaxx Comp Data Collection.",
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
	rootCmd.PersistentFlags().String("browser_user_agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.93 Safari/537.36", "browser user agent used for the data scraper")
	rootCmd.PersistentFlags().String("team_tag", "FOLLY", "team tag to track stats")
	rootCmd.PersistentFlags().Int("team_id", 1411729, "nitro type team id to cross check stats against")

	viper.BindPFlag("browser_user_agent", rootCmd.PersistentFlags().Lookup("browser_user_agent"))
	viper.BindPFlag("team_tag", rootCmd.PersistentFlags().Lookup("team_tag"))
	viper.BindPFlag("team_id", rootCmd.PersistentFlags().Lookup("team_id"))

	// Initialize cli
	cobra.OnInitialize(cli.InitConfig(rootCmd), func() {
		var err error
		logger, err = cli.CreateLogger()
		if err != nil {
			log.Fatalln("unable to setup logger")
		}
	})
}
