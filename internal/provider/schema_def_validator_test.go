// Copyright (c) Trifork

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSchemaDefValidator(t *testing.T) {
	tests := []struct {
		name          string
		schemaDef     string
		expectErr     bool
		errorContains string
	}{
		{
			name:      "null is skipped",
			schemaDef: "",
		},
		{
			name:      "basic string property",
			schemaDef: `{"summary":{"type":"string","description":"d"}}`,
		},
		{
			name:      "enum property",
			schemaDef: `{"sentiment":{"type":"enum","description":"d","enum":["a","b"]}}`,
		},
		{
			name:      "array property with nested items",
			schemaDef: `{"topics":{"type":"array","description":"d","items":{"type":"string","description":"d"}}}`,
		},
		{
			name:      "object property with nested properties",
			schemaDef: `{"author":{"type":"object","description":"d","properties":{"name":{"type":"string","description":"d"}}}}`,
		},
		{
			// The API's discriminator is "type", so enum values on a "string"
			// property are silently dropped instead of applied.
			name:          "enum on a string property is rejected",
			schemaDef:     `{"sentiment":{"type":"string","description":"d","enum":["a","b"]}}`,
			expectErr:     true,
			errorContains: `"enum"`,
		},
		{
			name:          "unknown type is rejected",
			schemaDef:     `{"summary":{"type":"str","description":"d"}}`,
			expectErr:     true,
			errorContains: "unsupported type",
		},
		{
			name:          "JSON Schema document shape is rejected",
			schemaDef:     `{"type":"object","properties":{"summary":{"type":"string"}},"required":["summary"]}`,
			expectErr:     true,
			errorContains: "type",
		},
		{
			name:          "missing type is rejected",
			schemaDef:     `{"summary":{"description":"d"}}`,
			expectErr:     true,
			errorContains: `"type"`,
		},
		{
			name:          "enum property without enum values is rejected",
			schemaDef:     `{"sentiment":{"type":"enum","description":"d"}}`,
			expectErr:     true,
			errorContains: `"enum"`,
		},
		{
			name:          "array property without items is rejected",
			schemaDef:     `{"topics":{"type":"array","description":"d"}}`,
			expectErr:     true,
			errorContains: `"items"`,
		},
		{
			name:          "nested items are validated too",
			schemaDef:     `{"topics":{"type":"array","description":"d","items":{"type":"string","description":"d","enum":["a"]}}}`,
			expectErr:     true,
			errorContains: `"enum"`,
		},
		{
			name:          "invalid JSON is rejected",
			schemaDef:     `{not json}`,
			expectErr:     true,
			errorContains: "valid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := types.StringNull()
			if tt.schemaDef != "" {
				val = types.StringValue(tt.schemaDef)
			}

			resp := &validator.StringResponse{}
			schemaDefValidator{}.ValidateString(
				context.Background(),
				validator.StringRequest{ConfigValue: val},
				resp,
			)

			if got := resp.Diagnostics.HasError(); got != tt.expectErr {
				t.Fatalf("HasError() = %v, want %v (diags: %v)", got, tt.expectErr, resp.Diagnostics)
			}
			if tt.errorContains != "" {
				if detail := resp.Diagnostics.Errors()[0].Detail(); !contains(detail, tt.errorContains) {
					t.Fatalf("error detail %q does not contain %q", detail, tt.errorContains)
				}
			}
		})
	}
}

func TestVariablesFromAPIInput(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]interface{}
		wantOK  bool
		wantLen int
	}{
		{name: "nil input", input: nil},
		{name: "missing key", input: map[string]interface{}{"other": 1}},
		{name: "explicit null", input: map[string]interface{}{"variables": nil}},
		{name: "empty list", input: map[string]interface{}{"variables": []interface{}{}}, wantOK: true},
		{
			name:    "list of names",
			input:   map[string]interface{}{"variables": []interface{}{"text", "tone"}},
			wantOK:  true,
			wantLen: 2,
		},
		{
			name:    "typed list of names",
			input:   map[string]interface{}{"variables": []string{"text"}},
			wantOK:  true,
			wantLen: 1,
		},
		{
			name:    "map keyed by name",
			input:   map[string]interface{}{"variables": map[string]interface{}{"text": "x"}},
			wantOK:  true,
			wantLen: 1,
		},
		{name: "non-string element", input: map[string]interface{}{"variables": []interface{}{1}}},
		{name: "unsupported shape", input: map[string]interface{}{"variables": "text"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got, ok := variablesFromAPIInput(tt.input, &diags)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if len(got) != tt.wantLen {
				t.Fatalf("variables = %v, want %d entries", got, tt.wantLen)
			}
			if diags.HasError() {
				t.Fatalf("unexpected error diagnostics: %v", diags.Errors())
			}
		})
	}
}
