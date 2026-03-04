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
	"github.com/spf13/cobra"
)

func NewAdminCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Administration commands",
		Long:  "Administration commands for managing users, roles, and policies",
	}

	cmd.AddCommand(newUserCommand())
	cmd.AddCommand(newRoleCommand())

	return cmd
}

func newUserCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "user",
		Short:   "Manage users",
		Aliases: []string{"users"},
	}

	cmd.AddCommand(newListUsersCommand())

	return cmd
}

func newListUsersCommand() *cobra.Command {
	var page, size int32

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List users",
		Long:  "List all users",
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

			_ = client
			_ = page
			_ = size

			return nil
		},
	}

	cmd.Flags().Int32VarP(&page, "page", "p", 1, "page number")
	cmd.Flags().Int32VarP(&size, "size", "s", 20, "page size")

	return cmd
}

func newRoleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "role",
		Short:   "Manage roles",
		Aliases: []string{"roles"},
	}

	cmd.AddCommand(newListRolesCommand())

	return cmd
}

func newListRolesCommand() *cobra.Command {
	var page, size int32

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List roles",
		Long:  "List all roles",
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

			_ = client
			_ = page
			_ = size

			return nil
		},
	}

	cmd.Flags().Int32VarP(&page, "page", "p", 1, "page number")
	cmd.Flags().Int32VarP(&size, "size", "s", 20, "page size")

	return cmd
}
