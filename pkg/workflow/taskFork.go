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
	"sync"

	"github.com/serverlessworkflow/sdk-go/v3/model"
	"go.temporal.io/sdk/workflow"
)

func forkTaskImpl(fork *model.ForkTask, task *model.TaskItem, workflowInst *Workflow) (TemporalWorkflowFunc, error) {
	childWorkflowName := GenerateChildWorkflowName("fork", task.Key)
	temporalWorkflows, err := workflowInst.workflowBuilder(fork.Fork.Branches, childWorkflowName)
	if err != nil {
		return nil, fmt.Errorf("error building forked workflow: %w", err)
	}

	n := len(temporalWorkflows)
	for _, t := range temporalWorkflows {
		n += len(t.Tasks)
	}

	return func(ctx workflow.Context, data *Variables, output map[string]OutputType) (*Future, error) {
		logger := workflow.GetLogger(ctx)
		logger.Debug("Forking a task", "isCompeting", fork.Fork.Compete)

		errs := make(chan error, n)
		var wg sync.WaitGroup

		for _, temporalWorkflow := range temporalWorkflows {
			for _, wf := range temporalWorkflow.Tasks {
				wg.Add(1)
				workflow.Go(ctx, func(ctx workflow.Context) {
					defer wg.Done()

					o := make(map[string]OutputType)

					future, err := wf.Task(ctx, data, o)
					if err != nil {
						logger.Error("Error handling Temporal task", "error", err, "task", wf.Key)
						errs <- err
						return
					}
					if future != nil {
						if err := future.Output(o); err != nil {
							errs <- err
							return
						}
					}
				})
			}
		}

		logger.Debug("Waiting for forked process to complete")
		wg.Wait()

		fmt.Println("Finished")

		select {
		case err := <-errs:
			return nil, err
		default:
		}

		return nil, nil
	}, nil
}
