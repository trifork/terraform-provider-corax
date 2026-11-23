// Copyright (c) Trifork

package coraxclient

import (
	"reflect"
	"testing"
)

func TestConvertConfigurationOmitsNullOptionalValues(t *testing.T) {
	actual := convertConfigurationFromGen(map[string]interface{}{
		"api_endpoint":         "https://example.services.ai.azure.com",
		"certificate_password": nil,
		"api_key":              "",
		"dimensions":           1024,
	})
	expected := map[string]string{
		"api_endpoint": "https://example.services.ai.azure.com",
		"api_key":      "",
		"dimensions":   "1024",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Unexpected configuration: %#v", actual)
	}
}
