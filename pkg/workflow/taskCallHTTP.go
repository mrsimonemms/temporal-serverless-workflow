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
	"bytes"
	"context"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"
	"time"

	"github.com/serverlessworkflow/sdk-go/v3/model"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

type CallHTTPResult struct {
	Body       string      `json:"body"`
	Headers    http.Header `json:"headers"`
	Method     string      `json:"method"`
	Status     string      `json:"status"`
	StatusCode int         `json:"statusCode"`
	URL        string      `json:"url"`
}

func (a *activities) CallHTTP(ctx context.Context, callHttp *model.CallHTTP, vars *Variables) (*CallHTTPResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Debug("Running call HTTP activity")

	body := bytes.NewBufferString(ParseVariables(bytes.NewBuffer(callHttp.With.Body).String(), vars))
	method := strings.ToUpper(ParseVariables(callHttp.With.Method, vars))
	url := ParseVariables(callHttp.With.Endpoint.String(), vars)

	logger.Debug("Making HTTP call", "method", method, "url", url)
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		logger.Error("Error making HTTP request", "method", method, "url", url, "error", err)
		return nil, fmt.Errorf("error making http request: %w", err)
	}

	for k, v := range callHttp.With.Headers {
		req.Header.Add(k, ParseVariables(v, vars))
	}

	q := req.URL.Query()
	for k, v := range callHttp.With.Query {
		q.Add(k, ParseVariables(v.(string), vars))
	}
	req.URL.RawQuery = q.Encode()

	// @todo(sje): configure the timeout
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error making HTTP call", "method", method, "url", url, "error", err)
		return nil, fmt.Errorf("error making http call: %w", err)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			logger.Error("Error closing body reader", "error", err)
		}
	}()

	bodyRes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading HTTP body", "method", method, "url", url, "error", err)
		return nil, fmt.Errorf("error reading http body: %w", err)
	}

	return &CallHTTPResult{
		Body:       string(bodyRes),
		Headers:    resp.Header,
		Method:     method,
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		URL:        url,
	}, err
}

func httpTaskImpl(task *model.CallHTTP, key string) TemporalWorkflowFunc {
	var a *activities

	return func(ctx workflow.Context, data *Variables, output map[string]OutputType) error {
		logger := workflow.GetLogger(ctx)

		logger.Debug("Calling HTTP endpoint")

		var result CallHTTPResult
		if err := workflow.ExecuteActivity(ctx, a.CallHTTP, task, data).Get(ctx, &result); err != nil {
			return fmt.Errorf("error calling http task: %w", err)
		}

		maps.Copy(output, map[string]OutputType{
			key: {
				Type: CallHTTPResultType,
				Data: result,
			},
		})

		return nil
	}
}
