// Copyright (c) Trifork

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// TestUpgradeSchemaDefFromRawState covers the schema_def state upgrade from the
// pre-1.0 dynamic representation (stored in state as a structured JSON value, or as
// a JSON string in normalized state) to the 1.x JSON-string representation.
func TestUpgradeSchemaDefFromRawState(t *testing.T) {
	tests := []struct {
		name      string
		rawState  string
		wantNull  bool
		wantValue string
	}{
		{
			name:      "structured object carried across as JSON text",
			rawState:  `{"id":"abc","schema_def":{"type":"object","properties":{"a":{"type":"string"}}}}`,
			wantValue: `{"type":"object","properties":{"a":{"type":"string"}}}`,
		},
		{
			name:      "structured array carried across as JSON text",
			rawState:  `{"schema_def":[{"type":"string"}]}`,
			wantValue: `[{"type":"string"}]`,
		},
		{
			name:      "already a JSON string is decoded, not double-encoded",
			rawState:  `{"schema_def":"{\"type\":\"object\"}"}`,
			wantValue: `{"type":"object"}`,
		},
		{
			name:     "explicit null maps to null",
			rawState: `{"schema_def":null}`,
			wantNull: true,
		},
		{
			name:     "absent attribute maps to null",
			rawState: `{"id":"abc"}`,
			wantNull: true,
		},
		{
			name:     "empty raw state maps to null",
			rawState: ``,
			wantNull: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := upgradeSchemaDefFromRawState([]byte(tt.rawState), &diags)

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags.Errors())
			}
			if tt.wantNull {
				if !got.IsNull() {
					t.Fatalf("expected null, got %q", got.ValueString())
				}
				return
			}
			if got.IsNull() {
				t.Fatalf("expected %q, got null", tt.wantValue)
			}
			if got.ValueString() != tt.wantValue {
				t.Fatalf("expected %q, got %q", tt.wantValue, got.ValueString())
			}
		})
	}
}
