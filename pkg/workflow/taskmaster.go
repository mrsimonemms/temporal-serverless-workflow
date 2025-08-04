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
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/serverlessworkflow/sdk-go/v3/impl/ctx"
	"github.com/serverlessworkflow/sdk-go/v3/model"
	"go.temporal.io/sdk/workflow"
)

type TemporalActivity struct{}

type TemporalWorkflow struct {
	fn func(workflow.Context)
}

type Temporal struct {
	Activities []*TemporalActivity
	Workflows  []*TemporalWorkflow
}

type Taskmaster struct {
	Workflow  *model.Workflow
	Context   context.Context
	RunnerCtx ctx.WorkflowContext
}

// Generate the Temporal workflows and activities
//
// Based on:
// @link https://github.com/serverlessworkflow/sdk-go/blob/592f31d64f24a3afd4d10040ae5a488f9600158b/impl/task_runner_do.go#L70
func (w *Workflow) Generate() (*Temporal, error) {
	temporal := &Temporal{
		Activities: make([]*TemporalActivity, 0),
		Workflows:  make([]*TemporalWorkflow, 0),
	}

	idx := 0
	currentTask := (*w.wf.Do)[idx]

	// There is always a primary workflow
	primaryWorkflow := func(ctx workflow.Context) {
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
	}

	for currentTask != nil {
		if err := w.SetTaskDef(currentTask); err != nil {
			return nil, fmt.Errorf("error setting task definition: %w", err)
		}
		if err := w.SetTaskReferenceFromName(currentTask.Key); err != nil {
			return nil, fmt.Errorf("error setting task reference: %w", err)
		}

		// @todo(sje): handle conditional if statement

		w.SetTaskStatus(currentTask.Key, ctx.PendingStatus)

		// Check if this task is a SwitchTask and handle it
		if task := currentTask.AsSwitchTask(); task != nil {
			log.Debug().Str("task", currentTask.Key).Msg("Registering switch task")
		}

		// Go to the next task
		idx, currentTask = w.wf.Do.Next(idx)
	}

	temporal.Workflows = append(temporal.Workflows, &TemporalWorkflow{
		fn: primaryWorkflow,
	})

	return temporal, nil
}
