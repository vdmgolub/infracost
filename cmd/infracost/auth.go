package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
)

func authCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with Infracost",
		Long:  "Authenticate with Infracost",
		Example: `  Login:

      infracost auth login`,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmds := []*cobra.Command{authLoginCmd(ctx)}
	cmd.AddCommand(cmds...)

	return cmd
}

func authLoginCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Infracost",
		Long:  "Authenticate with Infracost",
		Example: `  Login:

      infracost auth login`,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if ctx.Config.Credentials.APIKey != "" {
				fmt.Printf("You already have an Infracost API key saved in %s. We recommend using your same API key in all environments.\n", config.CredentialsFilePath())

				return nil
			}

			cmd.Printf("Redirecting to Infracost authentication page, go ahead and login, when you return back to this prompt the auth process should be completed.\n\n")

			auth := apiclient.AuthClient{Host: ctx.Config.DashboardAPIEndpoint}
			apiKey, err := auth.Login()
			if err != nil {
				cmd.Println(err)
				return err
			}

			ctx.Config.Credentials.APIKey = apiKey
			ctx.Config.Credentials.PricingAPIEndpoint = ctx.Config.PricingAPIEndpoint

			err = ctx.Config.Credentials.Save()
			if err != nil {
				return err
			}

			cmd.Printf("\nYour account has been authenticated. Infracost is now ready to be used.\n\n")
			fmt.Printf("The API key was saved to %s\n\n", config.CredentialsFilePath())

			return nil
		},
	}

	return cmd
}
