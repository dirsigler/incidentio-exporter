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
	"fmt"
	"log/slog"
)

// UserRole represents a base or custom role for a user.
type UserRole struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// User represents an incident.io user.
type User struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Role        string     `json:"role"` // Deprecated role field (viewer, responder, admin, owner)
	BaseRole    UserRole   `json:"base_role"`
	CustomRoles []UserRole `json:"custom_roles"`
	SlackUserID string     `json:"slack_user_id"`
}

// PaginationMeta represents pagination information for API responses.
type PaginationMeta struct {
	After    string `json:"after"`
	PageSize int    `json:"page_size"`
}

// UsersResponse holds the API response from the /v2/users endpoint.
type UsersResponse struct {
	Users          []User         `json:"users"`
	PaginationMeta PaginationMeta `json:"pagination_meta"`
}

// GetUsers retrieves all users from the incident.io API with pagination support.
func (c *Client) GetUsers(ctx context.Context) ([]User, error) {
	var allUsers []User
	pageCount := 0
	after := ""

	for {
		pageCount++
		path := fmt.Sprintf("/v2/users?page_size=%d", defaultPageSize)
		if after != "" {
			path += "&after=" + after
		}

		response := &UsersResponse{}
		if err := c.doRequest(ctx, path, response); err != nil {
			return nil, fmt.Errorf("failed to get users (page %d): %w", pageCount, err)
		}

		allUsers = append(allUsers, response.Users...)

		// Check if there are more pages
		if response.PaginationMeta.After == "" {
			break
		}
		after = response.PaginationMeta.After
	}

	slog.Info("completed fetching all users",
		"total_pages", pageCount,
		"total_users", len(allUsers),
	)

	return allUsers, nil
}
