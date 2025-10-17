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

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/pkg/cliutil"
	"github.com/olive-io/gflow/pkg/signalutil"
	"github.com/olive-io/gflow/pkg/version"
	"github.com/olive-io/gflow/runner"
)

func newRootCommand(stdout, stderr io.Writer) *cobra.Command {
	name := "gflow-runner"
	cfg := runner.NewConfig()
	app := &cobra.Command{
		Use:     name,
		Short:   "the runner component of gflow system",
		Version: version.ReleaseVersion(),
		PreRunE: func(cmd *cobra.Command, args []string) error {

			var err error
			cfgPath, _ := cmd.Flags().GetString("config")
			cfg, err = runner.FromConfigPath(cfgPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			if err = cfg.Init(); err != nil {
				return fmt.Errorf("init config: %w", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return startRunner(ctx, name, cfg)
		},
	}

	app.SetOut(stdout)
	app.SetErr(stderr)
	app.SetVersionTemplate(version.GetVersionTemplate())

	app.ResetFlags()
	flags := app.PersistentFlags()

	var configPath string
	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		configPath = filepath.Join(homeDir, ".olive", name+".toml")
	}

	flags.StringP("config", "C", configPath, "path to the configuration file")

	return app
}

func startRunner(ctx context.Context, name string, cfg *runner.Config) error {
	app, err := runner.New(name, cfg)
	if err != nil {
		return fmt.Errorf("create %s: %w", name, err)
	}

	ctx = signalutil.SetupSignalContext(ctx)

	// creates plugin factories
	factory := runner.NewTaskFactory(types.FlowNodeType_ServiceTask)
	//TODO: register plugins
	if err = app.SetupFactory(factory); err != nil {
		return fmt.Errorf("register %s: %w", name, err)
	}

	return app.Start(ctx)
}

func main() {
	cmd := newRootCommand(os.Stdout, os.Stderr)
	os.Exit(cliutil.Run(cmd))
}
