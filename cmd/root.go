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
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootOpts struct {
	FilePath string
	LogLevel string
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
		// Load the workflow file
		data, err := os.ReadFile(filepath.Clean(rootOpts.FilePath))
		if err != nil {
			log.Fatal().Err(err).Str("file", rootOpts.FilePath).Msg("Unable to load workflow file")
		}

		fmt.Println(string(data))
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
}
