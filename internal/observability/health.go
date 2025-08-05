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

package observability

import (
	"net"
	"net/http"

	"github.com/rs/zerolog/log"
	"go.temporal.io/sdk/client"
)

// Define a healthcheck listener
type healthcheck struct {
	client client.Client
}

func (h healthcheck) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	statusCode := http.StatusOK
	msg := "OK"

	if _, err := h.client.CheckHealth(r.Context(), &client.CheckHealthRequest{}); err != nil {
		statusCode = http.StatusServiceUnavailable
		msg = "Down"
	}

	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(msg))
}

// I'm not massively thrilled with using the logger, but it'll be ok
func NewHealthCheck(address string, c client.Client) {
	mux := http.NewServeMux()
	mux.Handle("/health", healthcheck{client: c})

	go func() {
		listener, err := net.Listen("tcp", address)
		if err != nil {
			log.Fatal().Err(err).Msg("Error creating listener")
		}

		defer func() {
			if err := listener.Close(); err != nil {
				log.Fatal().Err(err).Msg("Error closing health check connection")
			}
		}()

		//nolint:gosec
		if err := http.Serve(listener, mux); err != nil {
			log.Fatal().Err(err).Msg("Error serving health check connection")
		}
	}()
}
