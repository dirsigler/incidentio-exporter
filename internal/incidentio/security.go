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

import "strings"

// redactAPIKey masks the API key, showing only the last 4 characters for debugging.
func redactAPIKey(apiKey string) string {
	if len(apiKey) <= 4 {
		return "***"
	}
	return "***" + apiKey[len(apiKey)-4:]
}

// redactBearer removes the Bearer token from an Authorization header for safe logging.
func redactBearer(authHeader string) string {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "***"
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	return "Bearer " + redactAPIKey(token)
}
