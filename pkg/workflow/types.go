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
	"path/filepath"
	"strings"

	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/serverlessworkflow/sdk-go/v3/parser"
	"go.temporal.io/sdk/workflow"
)

type activities struct{}

type Future struct {
	ctx      workflow.Context
	key      string
	wfFuture workflow.Future
	// Output   func(workflow.Future, map[string]OutputType) error
}

func (f *Future) Output(output map[string]OutputType) error {
	logger := workflow.GetLogger(f.ctx)

	var result any
	if err := f.wfFuture.Get(f.ctx, &result); err != nil {
		logger.Error("Error calling http task", "error", err)
		return fmt.Errorf("error calling http task: %w", err)
	}

	maps.Copy(output, map[string]OutputType{
		f.key: {
			Type: CallHTTPResultType,
			Data: result,
		},
	})

	return nil
}

func NewFuture(ctx workflow.Context, key string, wfFuture workflow.Future) *Future {
	return &Future{
		ctx:      ctx,
		key:      key,
		wfFuture: wfFuture,
	}
}

type Workflow struct {
	data      []byte
	envPrefix string
	wf        *model.Workflow
}

type OutputType struct {
	Type ResultType `json:"type"`
	Data any        `json:"data"`
}

type Variables struct {
	Data map[string]any `json:"data"`
}

func (w *Workflow) Activities() *activities {
	return &activities{}
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

func LoadFromFile(file, envPrefix string) (*Workflow, error) {
	data, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		return nil, fmt.Errorf("error loading file: %w", err)
	}

	wf, err := parser.FromYAMLSource(data)
	if err != nil {
		return nil, fmt.Errorf("error loading yaml: %w", err)
	}

	return &Workflow{
		data:      data,
		envPrefix: strings.ToUpper(envPrefix),
		wf:        wf,
	}, nil
}
