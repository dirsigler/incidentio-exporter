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

import "fmt"

// APIError represents an error from the incident.io API.
type APIError struct {
	StatusCode int
	Endpoint   string
	Status     string
	Message    string
	Err        error
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("API error for %s: %s (HTTP %d %s): %v",
			e.Endpoint, e.Message, e.StatusCode, e.Status, e.Err)
	}
	return fmt.Sprintf("API error for %s: %s (HTTP %d %s)",
		e.Endpoint, e.Message, e.StatusCode, e.Status)
}

// Unwrap returns the wrapped error.
func (e *APIError) Unwrap() error {
	return e.Err
}

// IsRetryable returns true if the error is potentially retryable.
func (e *APIError) IsRetryable() bool {
	// 5xx errors and 429 (rate limit) are retryable
	return e.StatusCode >= 500 || e.StatusCode == 429
}

// NetworkError represents a network-level error.
type NetworkError struct {
	URL string
	Err error
}

// Error implements the error interface.
func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error for %s: %v", e.URL, e.Err)
}

// Unwrap returns the wrapped error.
func (e *NetworkError) Unwrap() error {
	return e.Err
}

// DecodeError represents a JSON decoding error.
type DecodeError struct {
	URL string
	Err error
}

// Error implements the error interface.
func (e *DecodeError) Error() string {
	return fmt.Sprintf("failed to decode JSON response from %s: %v", e.URL, e.Err)
}

// Unwrap returns the wrapped error.
func (e *DecodeError) Unwrap() error {
	return e.Err
}
