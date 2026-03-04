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
)

func NewSystemCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "System commands",
		Long:  "System commands for health check and runner management",
	}

	cmd.AddCommand(newPingCommand())

	return cmd
}

func newPingCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping gflow-server",
		Long:  "Check if gflow-server is healthy",
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
			err = client.Ping(ctx)
			if err != nil {
				return fmt.Errorf("ping failed: %w", err)
			}

			fmt.Println("pong")
			return nil
		},
	}

	return cmd
}
