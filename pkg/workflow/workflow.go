/*
 * Copyright 2025 Simon Emms <simon@simonemms.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package workflow

import (
	"fmt"
	"maps"
	"os"
	"strings"
	"time"

	"github.com/serverlessworkflow/sdk-go/v3/model"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflow) ToTemporalWorkflow(ctx workflow.Context) (map[string]OutputType, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Running workflow")

	timeout := time.Minute * 5
	if w.wf.Timeout != nil && w.wf.Timeout.Timeout != nil && w.wf.Timeout.Timeout.After != nil {
		timeout = ToDuration(w.wf.Timeout.Timeout.After)
	}

	logger.Debug("Setting workflow options", "StartToCloseTimeout", timeout)
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: timeout,
	})

	vars := &Variables{
		Data: make(map[string]any),
	}
	output := map[string]OutputType{}

	// Load in any envvars with the prefix
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if strings.HasPrefix(pair[0], w.envPrefix) {
			vars.Data[pair[0]] = pair[1]
		}
	}

	for _, task := range *w.wf.Do {
		logger.Info("Running task", "name", task.Key)

		if err := w.executeActivity(ctx, task, vars, output); err != nil {
			return nil, err
		}
	}

	return output, nil
}

func (w *Workflow) executeActivity(ctx workflow.Context, task *model.TaskItem, data *Variables, output map[string]OutputType) error {
	logger := workflow.GetLogger(ctx)
	var a *activities

	// Set data
	if set := task.AsSetTask(); set != nil {
		logger.Debug("Set data")

		for k, v := range set.Set {
			if s, ok := v.(string); ok {
				logger.Debug("Parsing value as string", "key", k)
				v = ParseVariables(s, data)
			}
			data.Data[ParseVariables(k, data)] = v
		}
	}

	// Have a little snooze
	if wait := task.AsWaitTask(); wait != nil {
		duration := ToDuration(wait.Wait)

		logger.Debug("Sleeping", "duration", duration.String())

		if err := workflow.Sleep(ctx, duration); err != nil {
			return fmt.Errorf("error sleeping: %w", err)
		}
	}

	// Call an HTTP endpoint
	if http := task.AsCallHTTPTask(); http != nil {
		logger.Debug("Calling HTTP endpoint")

		var result CallHTTPResult
		if err := workflow.ExecuteActivity(ctx, a.CallHTTP, http, data).Get(ctx, &result); err != nil {
			return fmt.Errorf("error calling http task: %w", err)
		}

		maps.Copy(output, map[string]OutputType{
			task.Key: {
				Type: CallHTTPResultType,
				Data: result,
			},
		})
	}

	return nil
}
