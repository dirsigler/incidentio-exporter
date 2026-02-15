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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecurity(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "redactAPIKey with normal API key",
			testFunc: func(t *testing.T) {
				apiKey := "sk_live_1234567890abcdef"
				result := redactAPIKey(apiKey)
				assert.Equal(t, "***cdef", result)
				assert.NotContains(t, result, "1234567890")
			},
		},
		{
			name: "redactAPIKey with short key (4 chars)",
			testFunc: func(t *testing.T) {
				apiKey := "test"
				result := redactAPIKey(apiKey)
				assert.Equal(t, "***", result)
			},
		},
		{
			name: "redactAPIKey with 3 char key shows nothing",
			testFunc: func(t *testing.T) {
				apiKey := "abc"
				result := redactAPIKey(apiKey)
				assert.Equal(t, "***", result)
			},
		},
		{
			name: "redactAPIKey with empty string",
			testFunc: func(t *testing.T) {
				apiKey := ""
				result := redactAPIKey(apiKey)
				assert.Equal(t, "***", result)
			},
		},
		{
			name: "redactAPIKey with single char",
			testFunc: func(t *testing.T) {
				apiKey := "x"
				result := redactAPIKey(apiKey)
				assert.Equal(t, "***", result)
			},
		},
		{
			name: "redactBearer with Bearer token",
			testFunc: func(t *testing.T) {
				token := "Bearer sk_live_1234567890abcdef"
				result := redactBearer(token)
				assert.Equal(t, "Bearer ***cdef", result)
				assert.Contains(t, result, "Bearer")
				assert.NotContains(t, result, "1234567890")
			},
		},
		{
			name: "redactBearer with plain token",
			testFunc: func(t *testing.T) {
				token := "sk_live_1234567890abcdef"
				result := redactBearer(token)
				assert.Equal(t, "***", result)
			},
		},
		{
			name: "redactBearer with Bearer and short token",
			testFunc: func(t *testing.T) {
				token := "Bearer xyz"
				result := redactBearer(token)
				assert.Equal(t, "Bearer ***", result)
			},
		},
		{
			name: "redactBearer with empty string",
			testFunc: func(t *testing.T) {
				token := ""
				result := redactBearer(token)
				assert.Equal(t, "***", result)
			},
		},
		{
			name: "redactBearer with Bearer only",
			testFunc: func(t *testing.T) {
				token := "Bearer "
				result := redactBearer(token)
				assert.Equal(t, "Bearer ***", result)
			},
		},
		{
			name: "redactBearer with multiple spaces after Bearer",
			testFunc: func(t *testing.T) {
				token := "Bearer   token123"
				result := redactBearer(token)
				// Should handle the first token part after Bearer
				assert.Contains(t, result, "Bearer")
				assert.Contains(t, result, "***")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}
