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
	"os"

	"github.com/spf13/cobra"

	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/clientgo"
)

func NewBpmnCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bpmn",
		Short: "BPMN commands",
		Long:  "BPMN commands for managing definitions and processes",
	}

	cmd.AddCommand(newDefinitionsCommand())
	cmd.AddCommand(newProcessCommand())

	return cmd
}

func newDefinitionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "definitions",
		Short:   "Manage BPMN definitions",
		Aliases: []string{"def", "definition"},
	}

	cmd.AddCommand(newDeployDefinitionCommand())
	cmd.AddCommand(newListDefinitionsCommand())
	cmd.AddCommand(newGetDefinitionCommand())

	return cmd
}

func newDeployDefinitionCommand() *cobra.Command {
	var description string

	cmd := &cobra.Command{
		Use:   "deploy <file>",
		Short: "Deploy a BPMN definition",
		Long:  "Deploy a BPMN definition from a file",
		Args:  cobra.ExactArgs(1),
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

			content, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}

			ctx := context.Background()
			definitions, err := client.DeployDefinitions(ctx, content, description, nil)
			if err != nil {
				return fmt.Errorf("deploy failed: %w", err)
			}

			fmt.Printf("Deployed: %s (v%d)\n", definitions.Uid, definitions.Version)
			return nil
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "definition description")

	return cmd
}

func newListDefinitionsCommand() *cobra.Command {
	var page, size int32

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List BPMN definitions",
		Long:  "List all BPMN definitions",
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

			req := &clientgo.ListDefinitionsRequest{
				Page: page,
				Size: size,
			}

			ctx := context.Background()
			list, total, err := client.ListDefinitions(ctx, req)
			if err != nil {
				return fmt.Errorf("list failed: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			printer := NewPrinter(GetOutputFormat(output))

			if GetOutputFormat(output) == OutputJSON {
				return printer.Print(list)
			}

			headers := []string{"UID", "NAME", "VERSION", "DESCRIPTION", "CREATED"}
			rows := make([][]string, len(list))
			for i, def := range list {
				rows[i] = []string{
					def.Uid,
					def.Name,
					fmt.Sprintf("%d", def.Version),
					def.Description,
					FormatTimestamp(def.CreateAt),
				}
			}

			fmt.Printf("Total: %d\n\n", total)
			return printer.PrintTable(headers, rows)
		},
	}

	cmd.Flags().Int32VarP(&page, "page", "p", 1, "page number")
	cmd.Flags().Int32VarP(&size, "size", "s", 20, "page size")

	return cmd
}

func newGetDefinitionCommand() *cobra.Command {
	var version uint64

	cmd := &cobra.Command{
		Use:   "get <uid>",
		Short: "Get a BPMN definition",
		Long:  "Get a BPMN definition by UID",
		Args:  cobra.ExactArgs(1),
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

			uid := args[0]

			ctx := context.Background()
			definitions, err := client.GetDefinitions(ctx, uid, version)
			if err != nil {
				return fmt.Errorf("get failed: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			printer := NewPrinter(GetOutputFormat(output))
			return printer.Print(definitions)
		},
	}

	cmd.Flags().Uint64VarP(&version, "version", "v", 0, "definition version (0 for latest)")

	return cmd
}

func newProcessCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "process",
		Short:   "Manage processes",
		Aliases: []string{"proc", "processes"},
	}

	cmd.AddCommand(newExecuteProcessCommand())
	cmd.AddCommand(newListProcessCommand())
	cmd.AddCommand(newGetProcessCommand())

	return cmd
}

func newExecuteProcessCommand() *cobra.Command {
	var name, definitionsUid string
	var definitionsVersion uint64
	var priority int64
	var mode string

	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute a process",
		Long:  "Execute a new process from a BPMN definition",
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

			transitionMode := types.TransitionMode_Simple
			if mode == "transition" {
				transitionMode = types.TransitionMode_Transition
			}

			req := &clientgo.ExecuteProcessRequest{
				Name:               name,
				DefinitionsUid:     definitionsUid,
				DefinitionsVersion: definitionsVersion,
				Priority:           priority,
				Mode:               transitionMode,
				Headers:            map[string]string{},
				Properties:         map[string]*types.Value{},
				DataObjects:        map[string]*types.Value{},
			}

			ctx := context.Background()
			process, err := client.ExecuteProcess(ctx, req)
			if err != nil {
				return fmt.Errorf("execute failed: %w", err)
			}

			fmt.Printf("Process started: %d\n", process.Id)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "process name")
	cmd.Flags().StringVarP(&definitionsUid, "uid", "u", "", "definitions UID")
	cmd.Flags().Uint64VarP(&definitionsVersion, "version", "v", 0, "definitions version")
	cmd.Flags().Int64VarP(&priority, "priority", "p", 0, "process priority")
	cmd.Flags().StringVarP(&mode, "mode", "m", "simple", "execution mode (simple|transition)")
	cmd.MarkFlagRequired("uid")

	return cmd
}

func newListProcessCommand() *cobra.Command {
	var page, size int32
	var definitionsUid string
	var definitionsVersion uint64

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List processes",
		Long:  "List all processes",
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

			req := &clientgo.ListProcessesRequest{
				Page:               page,
				Size:               size,
				DefinitionUID:      definitionsUid,
				DefinitionsVersion: definitionsVersion,
			}

			ctx := context.Background()
			processes, total, err := client.ListProcesses(ctx, req)
			if err != nil {
				return fmt.Errorf("list failed: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			printer := NewPrinter(GetOutputFormat(output))

			if GetOutputFormat(output) == OutputJSON {
				return printer.Print(processes)
			}

			headers := []string{"ID", "NAME", "STATUS", "STAGE", "STARTED"}
			rows := make([][]string, len(processes))
			for i, p := range processes {
				rows[i] = []string{
					fmt.Sprintf("%d", p.Id),
					p.Name,
					p.Status.String(),
					p.Stage.String(),
					FormatTimestamp(p.StartAt),
				}
			}

			fmt.Printf("Total: %d\n\n", total)
			return printer.PrintTable(headers, rows)
		},
	}

	cmd.Flags().Int32VarP(&page, "page", "p", 1, "page number")
	cmd.Flags().Int32VarP(&size, "size", "s", 20, "page size")
	cmd.Flags().StringVarP(&definitionsUid, "uid", "u", "", "filter by definitions UID")
	cmd.Flags().Uint64Var(&definitionsVersion, "version", 0, "filter by definitions version")

	return cmd
}

func newGetProcessCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a process",
		Long:  "Get a process by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var id int64
			fmt.Sscanf(args[0], "%d", &id)

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
			process, nodes, err := client.GetProcess(ctx, id)
			if err != nil {
				return fmt.Errorf("get failed: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			printer := NewPrinter(GetOutputFormat(output))

			if GetOutputFormat(output) == OutputJSON {
				return printer.Print(map[string]any{
					"process":    process,
					"activities": nodes,
				})
			}

			fmt.Printf("Process: %s (ID: %d)\n", process.Name, process.Id)
			fmt.Printf("Status: %s, Stage: %s\n", process.Status.String(), process.Stage.String())
			fmt.Printf("Started: %s\n", FormatTimestamp(process.StartAt))
			if process.EndAt > 0 {
				fmt.Printf("Ended: %s\n", FormatTimestamp(process.EndAt))
			}
			fmt.Printf("\nActivities (%d):\n", len(nodes))

			headers := []string{"NAME", "TYPE", "STATUS", "STAGE", "STARTED"}
			rows := make([][]string, len(nodes))
			for i, n := range nodes {
				rows[i] = []string{
					n.Name,
					n.FlowType.String(),
					n.Status.String(),
					n.Stage.String(),
					FormatTimestamp(n.StartTime),
				}
			}

			return printer.PrintTable(headers, rows)
		},
	}

	return cmd
}
