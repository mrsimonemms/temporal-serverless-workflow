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
	"path/filepath"

	"github.com/serverlessworkflow/sdk-go/v3/model"
	"github.com/serverlessworkflow/sdk-go/v3/parser"
)

type Workflow struct {
	data []byte
	wf   *model.Workflow
}

func (w *Workflow) ToTemporal() {}

func LoadFromFile(file string) (*Workflow, error) {
	data, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		return nil, fmt.Errorf("error loading file: %w", err)
	}

	wf, err := parser.FromYAMLSource(data)
	if err != nil {
		return nil, fmt.Errorf("error loading yaml: %w", err)
	}

	return &Workflow{
		data: data,
		wf:   wf,
	}, nil
}
