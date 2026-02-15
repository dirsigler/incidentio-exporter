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
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dirsigler/incidentio-exporter/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUsers(t *testing.T) {
	tests := []struct {
		name        string
		setupServer func() *httptest.Server
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, users []User)
	}{
		{
			name: "success with multiple users",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/v2/users?page_size=250", r.URL.Path+"?"+r.URL.RawQuery)
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"users": [
							{
								"id": "01FCNDV6P870EA6S7TK1DSYDG0",
								"name": "Lisa Karlin Curtis",
								"email": "lisa@incident.io",
								"role": "owner",
								"base_role": {
									"id": "01FCNDV6P870EA6S7TK1DSYDG1",
									"name": "Owner"
								},
								"custom_roles": [
									{
										"id": "01FCNDV6P870EA6S7TK1DSYDG2",
										"name": "Engineering Lead"
									}
								],
								"slack_user_id": "U02AYNF2XJM"
							},
							{
								"id": "01FCNDV6P870EA6S7TK1DSYDG3",
								"name": "John Doe",
								"email": "john@incident.io",
								"role": "responder",
								"base_role": {
									"id": "01FCNDV6P870EA6S7TK1DSYDG4",
									"name": "Responder"
								},
								"custom_roles": [],
								"slack_user_id": "U02AYNF2XJN"
							}
						],
						"pagination_meta": {
							"after": "",
							"page_size": 250
						}
					}`))
				}))
			},
			wantErr: false,
			checkResult: func(t *testing.T, users []User) {
				require.Len(t, users, 2)
				assert.Equal(t, "Lisa Karlin Curtis", users[0].Name)
				assert.Equal(t, "Owner", users[0].BaseRole.Name)
				assert.Len(t, users[0].CustomRoles, 1)
				assert.Equal(t, "Engineering Lead", users[0].CustomRoles[0].Name)
				assert.Equal(t, "John Doe", users[1].Name)
				assert.Equal(t, "Responder", users[1].BaseRole.Name)
				assert.Len(t, users[1].CustomRoles, 0)
			},
		},
		{
			name: "pagination with multiple pages",
			setupServer: func() *httptest.Server {
				page := 0
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					page++
					w.WriteHeader(http.StatusOK)
					if page == 1 {
						w.Write([]byte(`{
							"users": [
								{
									"id": "01FCNDV6P870EA6S7TK1DSYDG0",
									"name": "User 1",
									"email": "user1@incident.io",
									"role": "viewer",
									"base_role": {
										"id": "01FCNDV6P870EA6S7TK1DSYDG1",
										"name": "Viewer"
									},
									"custom_roles": [],
									"slack_user_id": "U02AYNF2XJM"
								}
							],
							"pagination_meta": {
								"after": "cursor-page-2",
								"page_size": 250
							}
						}`))
					} else {
						w.Write([]byte(`{
							"users": [
								{
									"id": "01FCNDV6P870EA6S7TK1DSYDG3",
									"name": "User 2",
									"email": "user2@incident.io",
									"role": "admin",
									"base_role": {
										"id": "01FCNDV6P870EA6S7TK1DSYDG4",
										"name": "Admin"
									},
									"custom_roles": [],
									"slack_user_id": "U02AYNF2XJN"
								}
							],
							"pagination_meta": {
								"after": "",
								"page_size": 250
							}
						}`))
					}
				}))
			},
			wantErr: false,
			checkResult: func(t *testing.T, users []User) {
				require.Len(t, users, 2)
				assert.Equal(t, "User 1", users[0].Name)
				assert.Equal(t, "User 2", users[1].Name)
			},
		},
		{
			name: "empty list",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"users": [],
						"pagination_meta": {
							"after": "",
							"page_size": 250
						}
					}`))
				}))
			},
			wantErr: false,
			checkResult: func(t *testing.T, users []User) {
				require.Len(t, users, 0)
			},
		},
		{
			name: "server error",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error": "internal server error"}`))
				}))
			},
			wantErr:     true,
			errContains: "500",
		},
		{
			name: "invalid JSON",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`invalid json`))
				}))
			},
			wantErr:     true,
			errContains: "failed to get users",
		},
		{
			name: "unauthorized",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`{"error": "unauthorized"}`))
				}))
			},
			wantErr:     true,
			errContains: "401",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupServer()
			defer server.Close()

			client := NewClient(config.IncidentIOConfig{
				Key:      "test-api-key",
				URL:      server.URL,
			})

			ctx := context.Background()
			users, err := client.GetUsers(ctx)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, users)
			}
		})
	}
}
