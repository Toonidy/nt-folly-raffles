package cli

import (
	"encoding/json"
	"fmt"
	"nt-folly-xmaxx-comp/pkg/nitrotype/clients"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch team stats (testing api).",
	Long:  "Fetch team stats (testing api).",
	Run: func(cmd *cobra.Command, args []string) {
		apiClient := clients.NewAPIClientBrowser(viper.GetString("browser_user_agent"))
		teamData, err := apiClient.GetTeam("FOLLY")
		if err != nil {
			logger.Error("failed to grab team stats", zap.Error(err))
			return
		}
		output, err := json.Marshal(teamData)
		if err != nil {
			logger.Error("failed marshal team stats", zap.Error(err))
			return
		}
		fmt.Println(string(output))
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)
}
