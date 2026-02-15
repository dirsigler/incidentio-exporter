// Copyright 2026 Dennis Irsigler
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package incidentio

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantError      string
		wantRetryable  bool
		wantUnwrap     bool
		checkUnwrapErr error
	}{
		{
			name: "APIError with 500 status",
			err: &APIError{
				StatusCode: 500,
				Endpoint:   "/v2/incidents",
				Status:     "Internal Server Error",
				Message:    "unexpected HTTP status",
			},
			wantError:     "API error for /v2/incidents: unexpected HTTP status (HTTP 500 Internal Server Error)",
			wantRetryable: true,
		},
		{
			name: "APIError with 429 rate limit",
			err: &APIError{
				StatusCode: 429,
				Endpoint:   "/v1/severities",
				Status:     "Too Many Requests",
				Message:    "rate limit exceeded",
			},
			wantError:     "API error for /v1/severities: rate limit exceeded (HTTP 429 Too Many Requests)",
			wantRetryable: true,
		},
		{
			name: "APIError with 401 unauthorized",
			err: &APIError{
				StatusCode: 401,
				Endpoint:   "/v1/incidents",
				Status:     "Unauthorized",
				Message:    "invalid credentials",
			},
			wantError:     "API error for /v1/incidents: invalid credentials (HTTP 401 Unauthorized)",
			wantRetryable: false,
		},
		{
			name: "APIError with wrapped error",
			err: &APIError{
				StatusCode: 503,
				Endpoint:   "/v2/incidents",
				Status:     "Service Unavailable",
				Message:    "service temporarily unavailable",
				Err:        errors.New("connection timeout"),
			},
			wantError:      "API error for /v2/incidents: service temporarily unavailable (HTTP 503 Service Unavailable): connection timeout",
			wantRetryable:  true,
			wantUnwrap:     true,
			checkUnwrapErr: errors.New("connection timeout"),
		},
		{
			name: "NetworkError",
			err: &NetworkError{
				URL: "https://api.incident.io/test",
				Err: errors.New("dial tcp: connection refused"),
			},
			wantError:      "network error for https://api.incident.io/test: dial tcp: connection refused",
			wantUnwrap:     true,
			checkUnwrapErr: errors.New("dial tcp: connection refused"),
		},
		{
			name: "DecodeError",
			err: &DecodeError{
				URL: "https://api.incident.io/v1/test",
				Err: errors.New("invalid character 'i' looking for beginning of object key string"),
			},
			wantError:      "failed to decode JSON response from https://api.incident.io/v1/test: invalid character 'i' looking for beginning of object key string",
			wantUnwrap:     true,
			checkUnwrapErr: errors.New("invalid character 'i' looking for beginning of object key string"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantError, tt.err.Error())

			if apiErr, ok := tt.err.(*APIError); ok {
				assert.Equal(t, tt.wantRetryable, apiErr.IsRetryable())
			}

			if tt.wantUnwrap {
				unwrappedErr := errors.Unwrap(tt.err)
				assert.NotNil(t, unwrappedErr)
				assert.Equal(t, tt.checkUnwrapErr.Error(), unwrappedErr.Error())
			}
		})
	}
}
