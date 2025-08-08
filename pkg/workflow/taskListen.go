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

func validateEventFilter(event *model.EventFilter) error {
	if event.With.ID == "" {
		return ErrUnsetListenTask
	}
	return nil
}

func listenConfigure(task *model.ListenTask, key string) (events []*model.EventFilter, isAll bool, err error) {
	isAll = false
	events = make([]*model.EventFilter, 0)

	if len(task.Listen.To.All) > 0 {
		isAll = true
		for k, i := range task.Listen.To.All {
			if err = validateEventFilter(i); err != nil {
				err = fmt.Errorf("%w: %s.%d", err, key, k)
				return events, isAll, err
			}
			events = append(events, i)
		}
	} else if len(task.Listen.To.Any) > 0 {
		for k, i := range task.Listen.To.Any {
			if err = validateEventFilter(i); err != nil {
				err = fmt.Errorf("%w: %s.%d", err, key, k)
				return events, isAll, err
			}
			events = append(events, i)
		}
	} else if task.Listen.To.One != nil {
		if err = validateEventFilter(task.Listen.To.One); err != nil {
			err = fmt.Errorf("%w: %s", err, key)
			return events, isAll, err
		}
		events = append(events, task.Listen.To.One)
	} else if task.Listen.To.Until != nil {
		err = fmt.Errorf("%w: listen.to.until", ErrUnsupportedTask)
		return events, isAll, err
	} else {
		err = ErrUnsetListenTask
		return events, isAll, err
	}

	return events, isAll, err
}

func listenTaskImpl(task *model.ListenTask, key string) (TemporalWorkflowFunc, error) {
	events, isAll, err := listenConfigure(task, key)
	if err != nil {
		return nil, err
	}

	return func(ctx workflow.Context, data *Variables, output map[string]OutputType) error {
		logger := workflow.GetLogger(ctx)
		logger.Debug("Registering listeners")

		isAllComplete := make([]bool, 0)
		isAnyComplete := false

		for i, event := range events {
			if isAll {
				isAllComplete = append(isAllComplete, false)
			}

			if err := workflow.SetUpdateHandler(ctx, event.With.ID, func(ctx workflow.Context, args HTTPData) error {
				fmt.Println("===")
				fmt.Printf("%+v\n", args)
				fmt.Println("===")

				logger.Debug("Listen event received", "event", event.With.ID)
				if isAll {
					isAllComplete[i] = true
				} else {
					isAnyComplete = true
				}

				return nil
			}); err != nil {
				logger.Error("Error setting update", "id", event.With.ID, "error", err)
				return fmt.Errorf("error setting update: %w", err)
			}
		}

		logger.Debug("Listening for updates")
		if err := workflow.Await(ctx, func() bool {
			if isAll {
				logger.Debug("Waiting for listener(s) to complete", "complete", isAllComplete)
				return SlicesEqual(isAllComplete, true)
			} else {
				logger.Debug("Waiting for listener to complete", "complete", isAnyComplete)
				return isAnyComplete
			}
		}); err != nil {
			logger.Error("Error waiting", "error", err)
			return fmt.Errorf("error waiting: %w", err)
		}

		return nil
	}, nil
}
