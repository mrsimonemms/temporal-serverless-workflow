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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/serverlessworkflow/sdk-go/v3/impl"
	"github.com/serverlessworkflow/sdk-go/v3/impl/ctx"
	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/serverlessworkflow/sdk-go/v3/parser"
)

type Workflow struct {
	data      []byte
	envPrefix string
	wf        *model.Workflow
	runnerCtx ctx.WorkflowContext
	ctx       context.Context
}

type OutputType struct {
	Type ResultType `json:"type"`
	Data any        `json:"data"`
}

type Variables struct {
	Data map[string]any `json:"data"`
}

func (w *Workflow) WorkflowName() string {
	return w.wf.Document.Name
}

// Validation of the schema is handled separately. This validates that there is
// nothing used we've not implemented. This should reduce over time.
func (w *Workflow) Validate() error {
	for _, task := range *w.wf.Do {
		if emit := task.AsEmitTask(); emit != nil {
			return fmt.Errorf("emit tasks are not supported")
		}
		if forTask := task.AsForTask(); forTask != nil {
			return fmt.Errorf("for tasks are not supported")
		}
		if fork := task.AsForkTask(); fork != nil {
			return fmt.Errorf("fork tasks are not supported")
		}
		if grpc := task.AsCallGRPCTask(); grpc != nil {
			return fmt.Errorf("grpc tasks are not supported")
		}
		if listen := task.AsListenTask(); listen != nil {
			return fmt.Errorf("listen tasks are not supported")
		}
		if openapi := task.AsCallOpenAPITask(); openapi != nil {
			return fmt.Errorf("openapi tasks are not supported")
		}
		if raise := task.AsRaiseTask(); raise != nil {
			return fmt.Errorf("raise tasks are not supported")
		}
		if run := task.AsRunTask(); run != nil {
			return fmt.Errorf("run tasks are not supported")
		}
		if switchTask := task.AsSwitchTask(); switchTask != nil {
			return fmt.Errorf("switch tasks are not supported")
		}
		if try := task.AsTryTask(); try != nil {
			return fmt.Errorf("try tasks are not supported")
		}
	}

	return nil
}

// AddLocalExprVars implements impl.TaskSupport.
func (w *Workflow) AddLocalExprVars(vars map[string]any) {
	w.runnerCtx.AddLocalExprVars(vars)
}

// CloneWithContext implements impl.TaskSupport.
func (w *Workflow) CloneWithContext(newCtx context.Context) impl.TaskSupport {
	clonedWfCtx := w.runnerCtx.Clone()

	ctxWithWf := ctx.WithWorkflowContext(newCtx, clonedWfCtx)

	return &Workflow{
		data:      w.data,
		envPrefix: w.envPrefix,
		wf:        w.wf,
		ctx:       ctxWithWf,
		runnerCtx: clonedWfCtx,
	}
}

// GetContext implements impl.TaskSupport.
func (w *Workflow) GetContext() context.Context {
	return w.ctx
}

// GetTaskReference implements impl.TaskSupport.
func (w *Workflow) GetTaskReference() string {
	return w.runnerCtx.GetTaskReference()
}

// GetWorkflowDef implements impl.TaskSupport.
func (w *Workflow) GetWorkflowDef() *model.Workflow {
	return w.wf
}

// RemoveLocalExprVars implements impl.TaskSupport.
func (w *Workflow) RemoveLocalExprVars(keys ...string) {
	w.runnerCtx.RemoveLocalExprVars(keys...)
}

// SetLocalExprVars implements impl.TaskSupport.
func (w *Workflow) SetLocalExprVars(vars map[string]any) {
	w.runnerCtx.SetLocalExprVars(vars)
}

// SetTaskDef implements impl.TaskSupport.
func (w *Workflow) SetTaskDef(task model.Task) error {
	return w.runnerCtx.SetTaskDef(task)
}

// SetTaskName implements impl.TaskSupport.
func (w *Workflow) SetTaskName(name string) {
	w.runnerCtx.SetTaskName(name)
}

// SetTaskRawInput implements impl.TaskSupport.
func (w *Workflow) SetTaskRawInput(input any) {
	w.runnerCtx.SetTaskRawInput(input)
}

// SetTaskRawOutput implements impl.TaskSupport.
func (w *Workflow) SetTaskRawOutput(output any) {
	w.runnerCtx.SetTaskRawOutput(output)
}

// SetTaskReferenceFromName implements impl.TaskSupport.
func (w *Workflow) SetTaskReferenceFromName(taskName string) error {
	ref, err := impl.GenerateJSONPointer(w.wf, taskName)
	if err != nil {
		return err
	}
	w.runnerCtx.SetTaskReference(ref)
	return nil
}

// SetTaskStartedAt implements impl.TaskSupport.
func (w *Workflow) SetTaskStartedAt(startedAt time.Time) {
	w.runnerCtx.SetTaskStartedAt(startedAt)
}

// SetTaskStatus implements impl.TaskSupport.
func (w *Workflow) SetTaskStatus(task string, status ctx.StatusPhase) {
	w.runnerCtx.SetTaskStatus(task, status)
}

// SetWorkflowInstanceCtx implements impl.TaskSupport.
func (w *Workflow) SetWorkflowInstanceCtx(value any) {
	w.runnerCtx.SetInstanceCtx(value)
}

func LoadFromFile(file, envPrefix string) (*Workflow, error) {
	data, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		return nil, fmt.Errorf("error loading file: %w", err)
	}

	wf, err := parser.FromYAMLSource(data)
	if err != nil {
		return nil, fmt.Errorf("error loading yaml: %w", err)
	}

	runnerCtx, err := ctx.NewWorkflowContext(wf)
	if err != nil {
		return nil, fmt.Errorf("error generating workflow context: %w", err)
	}

	return &Workflow{
		ctx:       ctx.WithWorkflowContext(context.Background(), runnerCtx),
		data:      data,
		envPrefix: strings.ToUpper(envPrefix),
		wf:        wf,
		runnerCtx: runnerCtx,
	}, nil
}
