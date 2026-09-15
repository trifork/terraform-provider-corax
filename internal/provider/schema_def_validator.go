// Copyright (c) Trifork

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// schemaDefAllowedKeys lists the keys the API accepts for each property type.
// The API discriminates on "type", so a key that does not belong to the given
// type is silently discarded server-side. Rejecting it here turns a confusing
// "Provider produced inconsistent result after apply" into an actionable error.
var schemaDefAllowedKeys = map[string][]string{
	"string":  {"type", "description"},
	"integer": {"type", "description"},
	"number":  {"type", "description"},
	"boolean": {"type", "description"},
	"enum":    {"type", "description", "enum"},
	"array":   {"type", "description", "items"},
	"object":  {"type", "description", "properties"},
}

// schemaDefRequiredKeys lists the type-specific key each property type needs on
// top of "type".
var schemaDefRequiredKeys = map[string]string{
	"enum":   "enum",
	"array":  "items",
	"object": "properties",
}

// schemaDefValidator validates the `schema_def` JSON string: a flat map of
// field name to property definition, where each property carries a `type`
// discriminator.
type schemaDefValidator struct{}

func (v schemaDefValidator) Description(ctx context.Context) string {
	return "Validates that schema_def is a JSON object mapping field names to property " +
		"definitions, each with a supported \"type\" and only the keys that type accepts."
}

func (v schemaDefValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v schemaDefValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var properties map[string]interface{}
	if err := json.Unmarshal([]byte(req.ConfigValue.ValueString()), &properties); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid schema_def",
			fmt.Sprintf("schema_def must be a valid JSON object mapping field names to property "+
				"definitions (use jsonencode()): %s", err),
		)
		return
	}

	for _, name := range sortedKeys(properties) {
		validateSchemaDefProperty(name, properties[name], req, resp)
	}
}

// validateSchemaDefProperty validates a single property definition, recursing
// into array items and object properties. `location` describes where the
// property sits, for error messages.
func validateSchemaDefProperty(location string, raw interface{}, req validator.StringRequest, resp *validator.StringResponse) {
	property, ok := raw.(map[string]interface{})
	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid schema_def",
			fmt.Sprintf("Property %q must be an object with a %q key, got %T.", location, "type", raw),
		)
		return
	}

	typeName, ok := property["type"].(string)
	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid schema_def",
			fmt.Sprintf("Property %q is missing a string %q key. schema_def is a flat map of field "+
				"name to property definition, not a JSON Schema document.", location, "type"),
		)
		return
	}

	allowed, ok := schemaDefAllowedKeys[typeName]
	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid schema_def",
			fmt.Sprintf("Property %q has unsupported type %q. Supported types: %s.",
				location, typeName, strings.Join(sortedKeys(schemaDefAllowedKeys), ", ")),
		)
		return
	}

	for _, key := range sortedKeys(property) {
		if !slicesContains(allowed, key) {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid schema_def",
				fmt.Sprintf("Property %q of type %q does not accept the key %q; it would be silently "+
					"dropped by the API. A type %q property accepts: %s.",
					location, typeName, key, typeName, strings.Join(allowed, ", ")),
			)
		}
	}

	if required, needs := schemaDefRequiredKeys[typeName]; needs {
		if _, present := property[required]; !present {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid schema_def",
				fmt.Sprintf("Property %q of type %q requires the key %q.", location, typeName, required),
			)
			return
		}
	}

	switch typeName {
	case "array":
		validateSchemaDefProperty(location+".items", property["items"], req, resp)
	case "object":
		nested, ok := property["properties"].(map[string]interface{})
		if !ok {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid schema_def",
				fmt.Sprintf("Property %q of type %q requires %q to be an object.", location, typeName, "properties"),
			)
			return
		}
		for _, name := range sortedKeys(nested) {
			validateSchemaDefProperty(location+"."+name, nested[name], req, resp)
		}
	}
}

// sortedKeys returns the map's keys in a stable order so diagnostics do not
// depend on Go's map iteration order.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func slicesContains(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}
