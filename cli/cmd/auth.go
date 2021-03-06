package cmd

import (
	"fmt"
	"net/url"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/uuid"
	"github.com/keptn/keptn/cli/utils"
	"github.com/keptn/keptn/cli/utils/credentialmanager"
	"github.com/spf13/cobra"
)

var endPoint *string
var apiToken *string

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticates the keptn CLI against a keptn installation.",
	Long: `Authenticates the keptn CLI against a keptn installation using an endpoint
and an API token. Both, the endpoint and API token are exposed during the keptn installation.
If the authentication is successful, the endpoint and the API token are stored in a password store. 

Example:
	keptn auth --endpoint=myendpoint.com --api-token=xyz`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.PrintLog("Starting to authenticate", utils.InfoLevel)

		source, _ := url.Parse("https://github.com/keptn/keptn/cli#auth")
		contentType := "application/json"
		var data interface{}
		event := cloudevents.Event{
			Context: cloudevents.EventContextV02{
				ID:          uuid.New().String(),
				Type:        "auth",
				Source:      types.URLRef{URL: *source},
				ContentType: &contentType,
			}.AsV02(),
			Data: data,
		}

		u, err := url.Parse(*endPoint)
		if err != nil {
			return err
		}

		authURL := *u
		authURL.Path = "auth"

		if !mocking {
			_, err = utils.Send(authURL, event, *apiToken)
			if err != nil {
				utils.PrintLog("Authentication was unsuccessful", utils.QuietLevel)
				return err
			}
			utils.PrintLog("Successfully authenticated", utils.InfoLevel)
			return credentialmanager.SetCreds(*u, *apiToken)
		}

		fmt.Println("skipping auth due to mocking flag set to true")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(authCmd)

	endPoint = authCmd.Flags().StringP("endpoint", "e", "", "The endpoint exposed by keptn")
	authCmd.MarkFlagRequired("endpoint")
	apiToken = authCmd.Flags().StringP("api-token", "a", "", "The API token provided by keptn")
	authCmd.MarkFlagRequired("api-token")
}
