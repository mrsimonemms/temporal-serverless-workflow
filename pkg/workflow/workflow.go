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
	"os"
	"strings"
	"time"

	"github.com/serverlessworkflow/sdk-go/v3/model"
	"go.temporal.io/sdk/workflow"
)

type TemporalWorkflowTask struct {
	Key  string
	Task TemporalWorkflowFunc
}

type TemporalWorkflowFunc func(ctx workflow.Context, data *Variables, output map[string]OutputType) error

type TemporalWorkflow struct {
	EnvPrefix string
	Name      string
	Timeout   time.Duration
	Tasks     []TemporalWorkflowTask
}

func (t *TemporalWorkflow) Workflow(ctx workflow.Context, data *Variables) (map[string]OutputType, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Running workflow")

	logger.Debug("Setting workflow options", "StartToCloseTimeout", t.Timeout)
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: t.Timeout,
	})

	vars := &Variables{
		Data: make(map[string]any),
	}
	output := map[string]OutputType{}

	// Load in any envvars with the prefix
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if strings.HasPrefix(pair[0], t.EnvPrefix) {
			vars.Data[pair[0]] = pair[1]
		}
	}

	for _, task := range t.Tasks {
		logger.Info("Running task", "name", task.Key)

		if err := task.Task(ctx, vars, output); err != nil {
			return nil, err
		}
	}

	return output, nil
}

func (w *Workflow) workflowBuilder(tasks *model.TaskList, name string) ([]*TemporalWorkflow, error) {
	wfs := make([]*TemporalWorkflow, 0)

	timeout := defaultWorkflowTimeout
	if w.wf.Timeout != nil && w.wf.Timeout.Timeout != nil && w.wf.Timeout.Timeout.After != nil {
		timeout = ToDuration(w.wf.Timeout.Timeout.After)
	}

	wf := &TemporalWorkflow{
		EnvPrefix: w.envPrefix,
		Name:      name,
		Tasks:     make([]TemporalWorkflowTask, 0),
		Timeout:   timeout,
	}

	// Iterate over the task list to build out our workflow(s)
	for _, item := range *tasks {
		var task TemporalWorkflowFunc
		var err error

		if fork := item.AsForkTask(); fork != nil {
			if task, err = forkTaskImpl(fork, item, w); err != nil {
				return nil, err
			}
		}

		if http := item.AsCallHTTPTask(); http != nil {
			task = httpTaskImpl(http, item.Key)
		}

		if set := item.AsSetTask(); set != nil {
			task = setTaskImpl(set)
		}

		if wait := item.AsWaitTask(); wait != nil {
			task = waitTaskImpl(wait)
		}

		wf.Tasks = append(wf.Tasks, TemporalWorkflowTask{
			Key:  item.Key,
			Task: task,
		})
	}

	// Add to the list of workflows
	wfs = append(wfs, wf)

	return wfs, nil
}

// This is the main workflow definition.
func (w *Workflow) BuildWorkflows() ([]*TemporalWorkflow, error) {
	wfs := make([]*TemporalWorkflow, 0)

	d, err := w.workflowBuilder(w.wf.Do, w.WorkflowName())
	if err != nil {
		return nil, fmt.Errorf("error building workflows: %w", err)
	}

	wfs = append(wfs, d...)
	return wfs, nil
}
