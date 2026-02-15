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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dirsigler/incidentio-exporter/internal/config"
)

func TestGetCustomFieldOptions(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.URL.Query().Get("custom_field_id") != "test-field-id" {
			t.Errorf("unexpected custom_field_id: %s", r.URL.Query().Get("custom_field_id"))
		}

		// Return mock response
		response := CustomFieldOptionsResponse{
			CustomFieldOptions: []CustomFieldOption{
				{
					ID:            "option-1",
					CustomFieldID: "test-field-id",
					Value:         "Option 1",
				},
				{
					ID:            "option-2",
					CustomFieldID: "test-field-id",
					Value:         "Option 2",
				},
			},
			PaginationMeta: struct {
				After    string `json:"after"`
				PageSize int    `json:"page_size"`
			}{
				After:    "",
				PageSize: 250,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with test server
	cfg := config.IncidentIOConfig{
		Key:      "test-key",
		URL:      server.URL,
	}
	client := NewClient(cfg)

	// Test GetCustomFieldOptions
	ctx := context.Background()
	result, err := client.GetCustomFieldOptions(ctx, "test-field-id")
	if err != nil {
		t.Fatalf("GetCustomFieldOptions failed: %v", err)
	}

	if len(result.CustomFieldOptions) != 2 {
		t.Errorf("expected 2 custom field options, got %d", len(result.CustomFieldOptions))
	}

	if result.CustomFieldOptions[0].Value != "Option 1" {
		t.Errorf("expected 'Option 1', got '%s'", result.CustomFieldOptions[0].Value)
	}
}

func TestGetCustomFieldOptionsPagination(t *testing.T) {
	pageCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageCount++

		var response CustomFieldOptionsResponse
		if pageCount == 1 {
			// First page
			response = CustomFieldOptionsResponse{
				CustomFieldOptions: []CustomFieldOption{
					{ID: "option-1", CustomFieldID: "test-field-id", Value: "Option 1"},
				},
				PaginationMeta: struct {
					After    string `json:"after"`
					PageSize int    `json:"page_size"`
				}{
					After:    "page2",
					PageSize: 1,
				},
			}
		} else {
			// Second page
			response = CustomFieldOptionsResponse{
				CustomFieldOptions: []CustomFieldOption{
					{ID: "option-2", CustomFieldID: "test-field-id", Value: "Option 2"},
				},
				PaginationMeta: struct {
					After    string `json:"after"`
					PageSize int    `json:"page_size"`
				}{
					After:    "",
					PageSize: 1,
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	cfg := config.IncidentIOConfig{
		Key:      "test-key",
		URL:      server.URL,
	}
	client := NewClient(cfg)

	ctx := context.Background()
	result, err := client.GetCustomFieldOptions(ctx, "test-field-id")
	if err != nil {
		t.Fatalf("GetCustomFieldOptions failed: %v", err)
	}

	if len(result.CustomFieldOptions) != 2 {
		t.Errorf("expected 2 custom field options across pages, got %d", len(result.CustomFieldOptions))
	}

	if pageCount != 2 {
		t.Errorf("expected 2 pages to be fetched, got %d", pageCount)
	}
}
