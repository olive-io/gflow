/*
Copyright 2025 The gflow Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/olive-io/gflow/clientgo"
)

func GetClient(cfg *Config) (*clientgo.Client, error) {
	return NewClient(cfg)
}

func parseConfig(cmd *cobra.Command) (*Config, error) {
	cfgFile, _ := cmd.Flags().GetString("config")
	return GetConfig(cfgFile)
}

func GetOutputFormat(output string) OutputFormat {
	return OutputFormat(output)
}

func NewApp() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gfcli",
		Short: "GFlow CLI - A command line tool for gflow-server",
		Long: `GFlow CLI is a command line tool for interacting with gflow-server.

It provides commands for managing BPMN definitions, processes, users, roles,
and other gflow resources.`,
	}

	cmd.PersistentFlags().StringP("config", "c", "", "config file path")
	cmd.PersistentFlags().StringP("output", "o", "simple", "output format (simple|json|table)")

	cmd.AddCommand(NewAuthCommand())
	cmd.AddCommand(NewBpmnCommand())
	cmd.AddCommand(NewAdminCommand())
	cmd.AddCommand(NewSystemCommand())

	return cmd
}

func NewAuthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
		Long:  "Authentication commands for login, logout, and user management",
	}

	cmd.AddCommand(newLoginCommand())
	cmd.AddCommand(newGetSelfCommand())

	return cmd
}

func newLoginCommand() *cobra.Command {
	var username, password string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to gflow-server",
		Long:  "Login to gflow-server with username and password",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := parseConfig(cmd)
			if err != nil {
				return err
			}
			client, err := GetClient(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := context.Background()
			token, err := client.Login(ctx, username, password)
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			fmt.Printf("Login successful!\n")
			fmt.Printf("Token: %s\n", token.Text)
			fmt.Printf("Expires: %s\n", FormatTimestamp(token.ExpireAt))

			return nil
		},
	}

	cmd.Flags().StringVarP(&username, "username", "u", "", "username")
	cmd.Flags().StringVarP(&password, "password", "p", "", "password")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")

	return cmd
}

func newGetSelfCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "self",
		Short: "Get current user info",
		Long:  "Get current authenticated user information",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := parseConfig(cmd)
			if err != nil {
				return err
			}
			client, err := GetClient(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := context.Background()
			user, role, policies, err := client.GetSelf(ctx)
			if err != nil {
				return fmt.Errorf("get self failed: %w", err)
			}

			fmt.Printf("User: %s (ID: %d)\n", user.Username, user.Id)
			fmt.Printf("Email: %s\n", user.Email)
			fmt.Printf("Role: %s\n", role.DisplayName)
			fmt.Printf("Policies: %d\n", len(policies))

			return nil
		},
	}

	return cmd
}
