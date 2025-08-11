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

	"github.com/serverlessworkflow/sdk-go/v3/model"
	"go.temporal.io/sdk/workflow"
)

// Wrap all set values in a SideEffect to allow for generated values
// to be safely used. This avoid non-deterministic errors, which are a
// pain in the arse in Temporalland
func setTaskValue(ctx workflow.Context, input string, data *Variables) (string, error) {
	var str string
	err := workflow.SideEffect(ctx, func(ctx workflow.Context) any {
		return MustParseVariables(input, data)
	}).Get(&str)
	if err != nil {
		return "", fmt.Errorf("unable to generate side effect value: %w", err)
	}

	return str, nil
}

func setTaskImpl(task *model.SetTask) TemporalWorkflowFunc {
	return func(ctx workflow.Context, data *Variables, output map[string]OutputType) error {
		logger := workflow.GetLogger(ctx)

		for k, v := range task.Set {
			var err error

			if rawStr, ok := v.(string); ok {
				logger.Debug("Parsing value as string", "key", k)
				v, err = setTaskValue(ctx, rawStr, data)
				if err != nil {
					return err
				}
			}

			k, err = setTaskValue(ctx, k, data)
			if err != nil {
				return err
			}

			// Set the data
			data.Data[k] = v
		}

		return nil
	}
}
