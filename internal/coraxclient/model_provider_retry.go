// Copyright (c) Trifork

package coraxclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const certificateReadinessTimeout = 10 * time.Minute

func isCertificateProvider(providerType string) bool {
	switch providerType {
	case "azure_service_principal_certificate", "azure_ai_service_principal_certificate", "azure_speech_service_principal_certificate":
		return true
	default:
		return false
	}
}

func isCertificateReadinessError(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusBadRequest {
		return false
	}
	var response struct {
		Detail string `json:"detail"`
	}
	if json.Unmarshal(apiErr.Body, &response) != nil {
		return false
	}
	// These validation errors occur before the backend saves a provider. Do not
	// retry arbitrary HTTP failures: a create may already have been committed.
	switch response.Detail {
	case "Certificate authentication is not configured. Ask your platform administrator to configure the certificate mount directory.",
		"No readable mounted certificate is available for this client ID. Ask your platform administrator to mount <client-id>.pem.":
		return true
	default:
		return false
	}
}

func retryCertificateReadiness(ctx context.Context, providerType string, operation func(context.Context) (*ModelProvider, error)) (*ModelProvider, error) {
	if !isCertificateProvider(providerType) {
		return operation(ctx)
	}
	ctx, cancel := context.WithTimeout(ctx, certificateReadinessTimeout)
	defer cancel()

	delay := time.Second
	var lastErr error
	for {
		if err := ctx.Err(); err != nil {
			if lastErr != nil {
				return nil, fmt.Errorf("waiting for backend certificate readiness: %w; last error: %w", err, lastErr)
			}
			return nil, err
		}
		result, err := operation(ctx)
		// Preserve the last readiness response if a later request is canceled.
		if err != nil && ctx.Err() != nil && lastErr != nil {
			continue
		}
		if !isCertificateReadinessError(err) {
			return result, err
		}
		lastErr = err
		tflog.Debug(ctx, "Retrying model provider request while backend certificate mount becomes ready", map[string]interface{}{
			"provider_type": providerType,
			"retry_delay":   delay.String(),
		})
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
		case <-timer.C:
		}
		delay = min(delay*2, 10*time.Second)
	}
}
