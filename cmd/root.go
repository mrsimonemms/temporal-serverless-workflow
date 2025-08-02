/*
Copyright © 2025 Simon Emms <simon@simonemms.com>

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
package cmd

import (
	"fmt"
	"os"

	"github.com/mrsimonemms/golang-helpers/temporal"
	tsw "github.com/mrsimonemms/temporal-serverless-workflow/pkg/workflow"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

var rootOpts struct {
	FilePath          string
	LogLevel          string
	TaskQueue         string
	TemporalAddress   string
	TemporalNamespace string
	Validate          bool
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "temporal-serverless-workflow",
	Version: Version,
	Short:   "Build Temporal workflows with Serverless Workflow",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		level, err := zerolog.ParseLevel(rootOpts.LogLevel)
		if err != nil {
			return err
		}
		zerolog.SetGlobalLevel(level)

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// The client and worker are heavyweight objects that should be created once per process.
		c, err := client.Dial(client.Options{
			HostPort:  rootOpts.TemporalAddress,
			Namespace: rootOpts.TemporalNamespace,
			Logger:    temporal.NewZerologHandler(&log.Logger),
		})
		if err != nil {
			log.Fatal().Err(err).Msg("Unable to create client")
		}
		defer c.Close()

		// Load the workflow file
		wf, err := tsw.LoadFromFile(rootOpts.FilePath)
		if err != nil {
			log.Fatal().Err(err).Msg("Error loading workflow")
		}

		if rootOpts.Validate {
			log.Debug().Msg("Running validation")
			if err := wf.Validate(); err != nil {
				log.Fatal().Err(err).Msg("Failed validation")
			}
		}

		w := worker.New(c, rootOpts.TaskQueue, worker.Options{})

		w.RegisterWorkflowWithOptions(wf.ToTemporalWorkflow, workflow.RegisterOptions{
			Name: wf.WorkflowName(),
		})
		w.RegisterActivity(wf.ToActivities())

		err = w.Run(worker.InterruptCh())
		if err != nil {
			log.Fatal().Err(err).Msg("Unable to start worker")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("WF")

	rootCmd.Flags().StringVarP(
		&rootOpts.FilePath,
		"file",
		"f",
		viper.GetString("file"),
		"Path to workflow file",
	)

	viper.SetDefault("log_level", zerolog.InfoLevel.String())
	rootCmd.PersistentFlags().StringVarP(
		&rootOpts.LogLevel,
		"log-level",
		"l",
		viper.GetString("log_level"),
		fmt.Sprintf("log level: %s", "Set log level"),
	)

	viper.SetDefault("task_queue", "serverless-workflow")
	rootCmd.Flags().StringVarP(
		&rootOpts.TaskQueue,
		"task-queue",
		"q",
		viper.GetString("task_queue"),
		"Task queue name",
	)

	viper.SetDefault("temporal_address", client.DefaultHostPort)
	rootCmd.Flags().StringVarP(
		&rootOpts.TemporalAddress,
		"temporal-address",
		"H",
		viper.GetString("temporal_address"),
		"Address of the Temporal server",
	)

	viper.SetDefault("temporal_namespace", client.DefaultNamespace)
	rootCmd.Flags().StringVarP(
		&rootOpts.TemporalNamespace,
		"temporal-namespace",
		"n",
		viper.GetString("temporal_namespace"),
		"Temporal namespace to use",
	)

	viper.SetDefault("validate", true)
	rootCmd.Flags().BoolVar(
		&rootOpts.Validate,
		"validate",
		viper.GetBool("validate"),
		"Run workflow validation",
	)
}
