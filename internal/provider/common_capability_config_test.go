// Copyright (c) Trifork

package provider

import (
	"context"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	api "terraform-provider-corax/internal/generated"
)

func TestCustomParametersToAPI(t *testing.T) {
	tests := []struct {
		name          string
		input         types.Dynamic
		expectedMap   map[string]interface{}
		expectError   bool
		errorContains string
	}{
		{
			name:        "nil dynamic value",
			input:       types.DynamicNull(),
			expectedMap: nil,
			expectError: false,
		},
		{
			name:        "unknown dynamic value",
			input:       types.DynamicUnknown(),
			expectedMap: nil,
			expectError: false,
		},
		{
			name:  "valid JSON string with mixed types",
			input: types.DynamicValue(types.StringValue(`{"key1":"value1","key2":123,"key3":true}`)),
			expectedMap: map[string]interface{}{
				"key1": "value1",
				"key2": float64(123),
				"key3": true,
			},
			expectError: false,
		},
		{
			name:          "invalid JSON string",
			input:         types.DynamicValue(types.StringValue(`{invalid json}`)),
			expectedMap:   nil,
			expectError:   true,
			errorContains: "custom_parameters was provided as a string, but it's not valid JSON",
		},
		{
			name:        "null string value",
			input:       types.DynamicValue(basetypes.NewStringNull()),
			expectedMap: nil,
			expectError: false,
		},
		{
			name:  "JSON string with nested objects",
			input: types.DynamicValue(types.StringValue(`{"nested":{"innerKey":"innerValue"},"topKey":"topValue"}`)),
			expectedMap: map[string]interface{}{
				"nested": map[string]interface{}{
					"innerKey": "innerValue",
				},
				"topKey": "topValue",
			},
			expectError: false,
		},
		{
			name:  "JSON string with array values",
			input: types.DynamicValue(types.StringValue(`{"items":["item1","item2","item3"]}`)),
			expectedMap: map[string]interface{}{
				"items": []interface{}{"item1", "item2", "item3"},
			},
			expectError: false,
		},
		{
			name:          "empty string value",
			input:         types.DynamicValue(types.StringValue("")),
			expectedMap:   nil,
			expectError:   true,
			errorContains: "custom_parameters was provided as a string, but it's not valid JSON",
		},
		{
			name:        "empty JSON string",
			input:       types.DynamicValue(types.StringValue(`{}`)),
			expectedMap: map[string]interface{}{},
			expectError: false,
		},
		{
			name: "HCL object with string values (user scenario)",
			input: types.DynamicValue(types.ObjectValueMust(
				map[string]attr.Type{
					"reasoning_effort": types.StringType,
					"verbosity":        types.StringType,
				},
				map[string]attr.Value{
					"reasoning_effort": types.StringValue("minimal"),
					"verbosity":        types.StringValue("low"),
				},
			)),
			expectedMap: map[string]interface{}{
				"reasoning_effort": "minimal",
				"verbosity":        "low",
			},
			expectError: false,
		},
		{
			name: "HCL object with mixed types",
			input: types.DynamicValue(types.ObjectValueMust(
				map[string]attr.Type{
					"temperature": types.Float64Type,
					"max_tokens":  types.Int64Type,
					"stream":      types.BoolType,
				},
				map[string]attr.Value{
					"temperature": types.Float64Value(0.7),
					"max_tokens":  types.Int64Value(1000),
					"stream":      types.BoolValue(true),
				},
			)),
			expectedMap: map[string]interface{}{
				"temperature": 0.7,
				"max_tokens":  int64(1000),
				"stream":      true,
			},
			expectError: false,
		},
		{
			// Terraform hands a bare HCL number literal to a Dynamic attribute
			// as NumberType, not Int64Type/Float64Type.
			name: "HCL object with number values as written in HCL",
			input: types.DynamicValue(types.ObjectValueMust(
				map[string]attr.Type{
					"retries":     types.NumberType,
					"temperature": types.NumberType,
				},
				map[string]attr.Value{
					"retries":     types.NumberValue(big.NewFloat(3)),
					"temperature": types.NumberValue(big.NewFloat(0.7)),
				},
			)),
			expectedMap: map[string]interface{}{
				"retries":     int64(3),
				"temperature": 0.7,
			},
			expectError: false,
		},
		{
			// Terraform hands a bare HCL list literal to a Dynamic attribute
			// as TupleType, not ListType.
			name: "HCL object with tuple value as written in HCL",
			input: types.DynamicValue(types.ObjectValueMust(
				map[string]attr.Type{
					"stop": types.TupleType{ElemTypes: []attr.Type{types.StringType, types.NumberType}},
				},
				map[string]attr.Value{
					"stop": types.TupleValueMust(
						[]attr.Type{types.StringType, types.NumberType},
						[]attr.Value{types.StringValue("END"), types.NumberValue(big.NewFloat(2))},
					),
				},
			)),
			expectedMap: map[string]interface{}{
				"stop": []interface{}{"END", int64(2)},
			},
			expectError: false,
		},
		{
			name: "HCL map with string values",
			input: types.DynamicValue(types.MapValueMust(
				types.StringType,
				map[string]attr.Value{
					"key1": types.StringValue("value1"),
					"key2": types.StringValue("value2"),
				},
			)),
			expectedMap: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			expectError: false,
		},
		{
			name: "null HCL object",
			input: types.DynamicValue(types.ObjectNull(
				map[string]attr.Type{
					"key": types.StringType,
				},
			)),
			expectedMap: nil,
			expectError: false,
		},
		{
			name: "HCL object with nested objects",
			input: types.DynamicValue(types.ObjectValueMust(
				map[string]attr.Type{
					"foo": types.ObjectType{AttrTypes: map[string]attr.Type{
						"bar": types.StringType,
					}},
				},
				map[string]attr.Value{
					"foo": types.ObjectValueMust(
						map[string]attr.Type{
							"bar": types.StringType,
						},
						map[string]attr.Value{
							"bar": types.StringValue("baz"),
						},
					),
				},
			)),
			expectedMap: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "baz",
				},
			},
			expectError: false,
		},
		{
			name: "HCL object with deeply nested structure",
			input: types.DynamicValue(types.ObjectValueMust(
				map[string]attr.Type{
					"level1": types.ObjectType{AttrTypes: map[string]attr.Type{
						"level2": types.ObjectType{AttrTypes: map[string]attr.Type{
							"level3": types.StringType,
							"value":  types.Int64Type,
						}},
					}},
				},
				map[string]attr.Value{
					"level1": types.ObjectValueMust(
						map[string]attr.Type{
							"level2": types.ObjectType{AttrTypes: map[string]attr.Type{
								"level3": types.StringType,
								"value":  types.Int64Type,
							}},
						},
						map[string]attr.Value{
							"level2": types.ObjectValueMust(
								map[string]attr.Type{
									"level3": types.StringType,
									"value":  types.Int64Type,
								},
								map[string]attr.Value{
									"level3": types.StringValue("deep"),
									"value":  types.Int64Value(42),
								},
							),
						},
					),
				},
			)),
			expectedMap: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"level3": "deep",
						"value":  int64(42),
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := customParametersToAPI(tt.input, &diags)

			if tt.expectError {
				if !diags.HasError() {
					t.Errorf("expected error but got none")
				}
				if tt.errorContains != "" {
					found := false
					for _, d := range diags.Errors() {
						if contains(d.Summary(), tt.errorContains) || contains(d.Detail(), tt.errorContains) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing %q, but got: %v", tt.errorContains, diags.Errors())
					}
				}
			} else {
				if diags.HasError() {
					t.Errorf("unexpected error: %v", diags.Errors())
				}
			}

			if !mapsEqual(result, tt.expectedMap) {
				t.Errorf("result = %v, want %v", result, tt.expectedMap)
			}
		})
	}
}

func TestCustomParametersAPIToTerraform(t *testing.T) {
	tests := []struct {
		name          string
		input         map[string]interface{}
		expectNull    bool
		expectError   bool
		validateValue func(t *testing.T, result types.Dynamic)
	}{
		{
			name:       "nil input",
			input:      nil,
			expectNull: true,
		},
		{
			name:       "empty map",
			input:      map[string]interface{}{},
			expectNull: false,
			validateValue: func(t *testing.T, result types.Dynamic) {
				if result.IsNull() {
					t.Error("expected non-null result for empty map")
				}
				underlyingVal := result.UnderlyingValue()
				objVal, ok := underlyingVal.(types.Object)
				if !ok {
					t.Errorf("expected underlying value to be types.Object, got %T", underlyingVal)
					return
				}
				attrs := objVal.Attributes()
				if len(attrs) != 0 {
					t.Errorf("expected empty object, got %d attributes", len(attrs))
				}
			},
		},
		{
			name: "map with string values",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			expectNull: false,
			validateValue: func(t *testing.T, result types.Dynamic) {
				if result.IsNull() {
					t.Error("expected non-null result")
				}
				underlyingVal := result.UnderlyingValue()
				objVal, ok := underlyingVal.(types.Object)
				if !ok {
					t.Errorf("expected underlying value to be types.Object, got %T", underlyingVal)
					return
				}
				attrs := objVal.Attributes()
				if len(attrs) != 2 {
					t.Errorf("expected 2 attributes, got %d", len(attrs))
				}

				key1Val, ok := attrs["key1"].(types.String)
				if !ok || key1Val.ValueString() != "value1" {
					t.Errorf("expected key1=value1")
				}

				key2Val, ok := attrs["key2"].(types.String)
				if !ok || key2Val.ValueString() != "value2" {
					t.Errorf("expected key2=value2")
				}
			},
		},
		{
			name: "map with mixed value types",
			input: map[string]interface{}{
				"stringKey": "test",
				"boolKey":   true,
				"intKey":    42,
				"floatKey":  3.14,
			},
			expectNull: false,
			validateValue: func(t *testing.T, result types.Dynamic) {
				if result.IsNull() {
					t.Error("expected non-null result")
				}
				underlyingVal := result.UnderlyingValue()
				objVal, ok := underlyingVal.(types.Object)
				if !ok {
					t.Errorf("expected underlying value to be types.Object, got %T", underlyingVal)
					return
				}
				attrs := objVal.Attributes()
				if len(attrs) != 4 {
					t.Errorf("expected 4 attributes, got %d", len(attrs))
				}

				stringVal, ok := attrs["stringKey"].(types.String)
				if !ok || stringVal.ValueString() != "test" {
					t.Errorf("expected stringKey=test")
				}

				boolVal, ok := attrs["boolKey"].(types.Bool)
				if !ok || !boolVal.ValueBool() {
					t.Errorf("expected boolKey=true")
				}

				intVal, ok := attrs["intKey"].(types.Int64)
				if !ok || intVal.ValueInt64() != 42 {
					t.Errorf("expected intKey=42")
				}

				floatVal, ok := attrs["floatKey"].(types.Float64)
				if !ok || floatVal.ValueFloat64() != 3.14 {
					t.Errorf("expected floatKey=3.14")
				}
			},
		},
		{
			name: "map with nested structure",
			input: map[string]interface{}{
				"nested": map[string]interface{}{
					"innerKey": "innerValue",
				},
				"array": []interface{}{"item1", "item2"},
			},
			expectNull: false,
			validateValue: func(t *testing.T, result types.Dynamic) {
				if result.IsNull() {
					t.Error("expected non-null result")
				}
				underlyingVal := result.UnderlyingValue()
				objVal, ok := underlyingVal.(types.Object)
				if !ok {
					t.Errorf("expected underlying value to be types.Object, got %T", underlyingVal)
					return
				}
				attrs := objVal.Attributes()
				if len(attrs) != 2 {
					t.Errorf("expected 2 attributes, got %d", len(attrs))
				}

				nestedVal, ok := attrs["nested"].(types.Object)
				if !ok {
					t.Errorf("expected nested to be types.Object, got %T", attrs["nested"])
					return
				}
				nestedAttrs := nestedVal.Attributes()
				innerKeyVal, ok := nestedAttrs["innerKey"].(types.String)
				if !ok || innerKeyVal.ValueString() != "innerValue" {
					t.Errorf("expected nested.innerKey=innerValue")
				}

				arrayVal, ok := attrs["array"].(types.List)
				if !ok {
					t.Errorf("expected array to be types.List, got %T", attrs["array"])
					return
				}
				if len(arrayVal.Elements()) != 2 {
					t.Errorf("expected array to have 2 elements, got %d", len(arrayVal.Elements()))
				}
			},
		},
		{
			name: "should return Object not String (reproduces Terraform error)",
			input: map[string]interface{}{
				"reasoning": "minimal",
				"verbosity": "low",
			},
			expectNull: false,
			validateValue: func(t *testing.T, result types.Dynamic) {
				if result.IsNull() {
					t.Error("expected non-null result")
				}
				underlyingVal := result.UnderlyingValue()

				// The underlying value should be types.Object, NOT types.String
				// If it's types.String, Terraform will see it as a string type and error with:
				// "attribute custom_parameters: object required, but have string"
				objVal, ok := underlyingVal.(types.Object)
				if !ok {
					t.Errorf("expected underlying value to be types.Object, got %T (this is the bug!)", underlyingVal)
					return
				}

				// Verify the object has the correct attributes
				attrs := objVal.Attributes()
				if len(attrs) != 2 {
					t.Errorf("expected 2 attributes, got %d", len(attrs))
				}

				// Check that the values are correct
				reasoningVal, ok := attrs["reasoning"]
				if !ok {
					t.Error("expected 'reasoning' attribute")
				} else {
					reasoningStr, ok := reasoningVal.(types.String)
					if !ok {
						t.Errorf("expected reasoning to be types.String, got %T", reasoningVal)
					} else if reasoningStr.ValueString() != "minimal" {
						t.Errorf("expected reasoning='minimal', got %q", reasoningStr.ValueString())
					}
				}

				verbosityVal, ok := attrs["verbosity"]
				if !ok {
					t.Error("expected 'verbosity' attribute")
				} else {
					verbosityStr, ok := verbosityVal.(types.String)
					if !ok {
						t.Errorf("expected verbosity to be types.String, got %T", verbosityVal)
					} else if verbosityStr.ValueString() != "low" {
						t.Errorf("expected verbosity='low', got %q", verbosityStr.ValueString())
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := customParametersAPIToTerraform(tt.input, &diags)

			if tt.expectError {
				if !diags.HasError() {
					t.Error("expected error but got none")
				}
			} else {
				if diags.HasError() {
					t.Errorf("unexpected error: %v", diags.Errors())
				}
			}

			if tt.expectNull {
				if !result.IsNull() {
					t.Error("expected null result")
				}
			} else {
				if tt.validateValue != nil {
					tt.validateValue(t, result)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func mapsEqual(a, b map[string]interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || !valuesEqual(v, bv) {
			return false
		}
	}
	return true
}

func valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch aVal := a.(type) {
	case map[string]interface{}:
		bVal, ok := b.(map[string]interface{})
		if !ok {
			return false
		}
		return mapsEqual(aVal, bVal)
	case []interface{}:
		bVal, ok := b.([]interface{})
		if !ok {
			return false
		}
		if len(aVal) != len(bVal) {
			return false
		}
		for i := range aVal {
			if !valuesEqual(aVal[i], bVal[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}

// configObject builds a config-shaped object from a partial set of attributes,
// filling every attribute that is not supplied with a null of the right type.
func configObject(t *testing.T, attrs map[string]attr.Value) types.Object {
	t.Helper()

	full := map[string]attr.Value{
		"temperature":       types.Float64Null(),
		"blob_config":       types.ObjectNull(blobConfigAttributeTypes()),
		"data_retention":    types.ObjectNull(dataRetentionAttributeTypes()),
		"content_tracing":   types.BoolNull(),
		"custom_parameters": types.DynamicNull(),
		"mcp_server_ids":    types.ListNull(types.StringType),
	}
	for name, value := range attrs {
		if _, ok := full[name]; !ok {
			t.Fatalf("unknown config attribute %q", name)
		}
		full[name] = value
	}

	obj, diags := types.ObjectValue(capabilityConfigAttributeTypes(), full)
	if diags.HasError() {
		t.Fatalf("building config object: %v", diags)
	}
	return obj
}

func TestPreserveUnknownsFromState(t *testing.T) {
	dataRetention := types.ObjectValueMust(dataRetentionAttributeTypes(), map[string]attr.Value{
		"type":  types.StringValue("timed"),
		"hours": types.Int64Value(24),
	})

	t.Run("unknown nested attributes take the value from state", func(t *testing.T) {
		// Given a plan where Terraform marked the computed attributes unknown
		// because the configuration only sets temperature.
		plan := configObject(t, map[string]attr.Value{
			"temperature":     types.Float64Value(0.7),
			"content_tracing": types.BoolUnknown(),
			"data_retention":  types.ObjectUnknown(dataRetentionAttributeTypes()),
		})
		state := configObject(t, map[string]attr.Value{
			"temperature":     types.Float64Value(0.2),
			"content_tracing": types.BoolValue(true),
			"data_retention":  dataRetention,
		})
		cfg := configObject(t, map[string]attr.Value{"temperature": types.Float64Value(0.7)})

		// When the unknowns are resolved against state.
		got, diags := preserveUnknownsFromState(context.Background(), plan, state, cfg)

		// Then the configured value wins and the computed ones come from state.
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		want := configObject(t, map[string]attr.Value{
			"temperature":     types.Float64Value(0.7),
			"content_tracing": types.BoolValue(true),
			"data_retention":  dataRetention,
		})
		if !got.Equal(want) {
			t.Errorf("preserveUnknownsFromState() = %v, want %v", got, want)
		}
	})

	t.Run("null state attributes stay null instead of unknown", func(t *testing.T) {
		// Given state where the API never returned a data retention block.
		plan := configObject(t, map[string]attr.Value{
			"temperature":    types.Float64Value(0.7),
			"data_retention": types.ObjectUnknown(dataRetentionAttributeTypes()),
		})
		state := configObject(t, map[string]attr.Value{"temperature": types.Float64Value(0.2)})
		cfg := configObject(t, map[string]attr.Value{"temperature": types.Float64Value(0.7)})

		// When the unknowns are resolved against state.
		got, diags := preserveUnknownsFromState(context.Background(), plan, state, cfg)

		// Then the plan is null there too, so the plan stays empty.
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if dr := got.Attributes()["data_retention"]; !dr.IsNull() {
			t.Errorf("data_retention = %v, want null", dr)
		}
	})

	t.Run("unknown values inside nested objects take the value from state", func(t *testing.T) {
		// Given a blob_config whose computed attributes are unknown in the plan.
		planBlob := types.ObjectValueMust(blobConfigAttributeTypes(), map[string]attr.Value{
			"max_file_size_mb":   types.Int64Value(30),
			"max_blobs":          types.Int64Unknown(),
			"allowed_mime_types": types.ListUnknown(types.StringType),
		})
		stateBlob := types.ObjectValueMust(blobConfigAttributeTypes(), map[string]attr.Value{
			"max_file_size_mb":   types.Int64Value(20),
			"max_blobs":          types.Int64Value(10),
			"allowed_mime_types": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("image/png")}),
		})
		cfgBlob := types.ObjectValueMust(blobConfigAttributeTypes(), map[string]attr.Value{
			"max_file_size_mb":   types.Int64Value(30),
			"max_blobs":          types.Int64Null(),
			"allowed_mime_types": types.ListNull(types.StringType),
		})

		// When the unknowns are resolved against state.
		got, diags := preserveUnknownsFromState(context.Background(),
			configObject(t, map[string]attr.Value{"blob_config": planBlob}),
			configObject(t, map[string]attr.Value{"blob_config": stateBlob}),
			configObject(t, map[string]attr.Value{"blob_config": cfgBlob}))

		// Then only the configured attribute differs from state.
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		wantBlob := types.ObjectValueMust(blobConfigAttributeTypes(), map[string]attr.Value{
			"max_file_size_mb":   types.Int64Value(30),
			"max_blobs":          types.Int64Value(10),
			"allowed_mime_types": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("image/png")}),
		})
		if blob := got.Attributes()["blob_config"]; !blob.Equal(wantBlob) {
			t.Errorf("blob_config = %v, want %v", blob, wantBlob)
		}
	})

	t.Run("attributes unknown in configuration stay unknown", func(t *testing.T) {
		// Given a configuration value that depends on another resource.
		plan := configObject(t, map[string]attr.Value{"content_tracing": types.BoolUnknown()})
		state := configObject(t, map[string]attr.Value{"content_tracing": types.BoolValue(true)})
		cfg := configObject(t, map[string]attr.Value{"content_tracing": types.BoolUnknown()})

		// When the unknowns are resolved against state.
		got, diags := preserveUnknownsFromState(context.Background(), plan, state, cfg)

		// Then it must stay unknown, since the configured value decides it.
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if ct := got.Attributes()["content_tracing"]; !ct.IsUnknown() {
			t.Errorf("content_tracing = %v, want unknown", ct)
		}
	})
}

// TestCapabilityConfigAPItoModelMcpServerIds covers the API returning an empty
// mcp_server_ids list for a config that never set one. Mapping [] to an empty
// list instead of null makes Terraform reject the apply with "produced an
// unexpected new value: .config.mcp_server_ids: was null, but now
// cty.ListValEmpty(cty.String)".
func TestCapabilityConfigAPItoModelMcpServerIds(t *testing.T) {
	tests := []struct {
		name     string
		apiIDs   []string
		wantNull bool
		wantIDs  []string
	}{
		{name: "nil from API is null", apiIDs: nil, wantNull: true},
		{name: "empty list from API is null", apiIDs: []string{}, wantNull: true},
		{
			name:    "populated list from API is preserved",
			apiIDs:  []string{"11111111-1111-1111-1111-111111111111"},
			wantIDs: []string{"11111111-1111-1111-1111-111111111111"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			obj := capabilityConfigAPItoModel(context.Background(), &api.CapabilityConfig{McpServerIds: tt.apiIDs}, &diags)
			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags.Errors())
			}

			got, ok := obj.Attributes()["mcp_server_ids"].(types.List)
			if !ok {
				t.Fatalf("mcp_server_ids is not a List: %T", obj.Attributes()["mcp_server_ids"])
			}
			if tt.wantNull {
				if !got.IsNull() {
					t.Fatalf("mcp_server_ids = %v, want null", got)
				}
				return
			}
			var ids []string
			got.ElementsAs(context.Background(), &ids, false)
			if len(ids) != len(tt.wantIDs) || (len(ids) > 0 && ids[0] != tt.wantIDs[0]) {
				t.Fatalf("mcp_server_ids = %v, want %v", ids, tt.wantIDs)
			}
		})
	}
}

// TestCapabilityConfigValidatorContentTracing covers the cross-field rule the
// Corax API enforces server-side: timed data retention forces content_tracing
// to false. Without a plan-time error the apply fails much later with
// "produced an unexpected new value: .config.content_tracing: was cty.True,
// but now cty.False".
func TestCapabilityConfigValidatorContentTracing(t *testing.T) {
	newConfig := func(contentTracing attr.Value, retentionType string) types.Object {
		retention := types.ObjectNull(dataRetentionAttributeTypes())
		if retentionType != "" {
			retention = types.ObjectValueMust(dataRetentionAttributeTypes(), map[string]attr.Value{
				"type":  types.StringValue(retentionType),
				"hours": types.Int64Value(24),
			})
		}
		return configObject(t, map[string]attr.Value{
			"content_tracing": contentTracing,
			"data_retention":  retention,
		})
	}

	tests := []struct {
		name      string
		config    types.Object
		expectErr bool
	}{
		{
			name:      "content_tracing true with timed retention is rejected",
			config:    newConfig(types.BoolValue(true), "timed"),
			expectErr: true,
		},
		{
			name:   "content_tracing false with timed retention is allowed",
			config: newConfig(types.BoolValue(false), "timed"),
		},
		{
			name:   "content_tracing true with infinite retention is allowed",
			config: newConfig(types.BoolValue(true), "infinite"),
		},
		{
			name:   "content_tracing true without retention is allowed",
			config: newConfig(types.BoolValue(true), ""),
		},
		{
			name:   "content_tracing unset with timed retention is allowed",
			config: newConfig(types.BoolNull(), "timed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &validator.ObjectResponse{}
			contentTracingRetentionValidator{}.ValidateObject(
				context.Background(),
				validator.ObjectRequest{ConfigValue: tt.config},
				resp,
			)
			if got := resp.Diagnostics.HasError(); got != tt.expectErr {
				t.Fatalf("HasError() = %v, want %v (diags: %v)", got, tt.expectErr, resp.Diagnostics)
			}
		})
	}
}
