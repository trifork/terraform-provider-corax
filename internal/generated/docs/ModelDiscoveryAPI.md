# \ModelDiscoveryAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ValidateProviderV1ModelDiscoveryValidatePost**](ModelDiscoveryAPI.md#ValidateProviderV1ModelDiscoveryValidatePost) | **Post** /v1/model-discovery/validate | Validate Provider



## ValidateProviderV1ModelDiscoveryValidatePost

> ProviderValidationResponse ValidateProviderV1ModelDiscoveryValidatePost(ctx).Request(request).Execute()

Validate Provider



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID/api"
)

func main() {
	request := *openapiclient.NewRequest(*openapiclient.NewOpenAIConfiguration("ApiKey_example"), "ProviderType_example") // Request | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ModelDiscoveryAPI.ValidateProviderV1ModelDiscoveryValidatePost(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ModelDiscoveryAPI.ValidateProviderV1ModelDiscoveryValidatePost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ValidateProviderV1ModelDiscoveryValidatePost`: ProviderValidationResponse
	fmt.Fprintf(os.Stdout, "Response from `ModelDiscoveryAPI.ValidateProviderV1ModelDiscoveryValidatePost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiValidateProviderV1ModelDiscoveryValidatePostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**Request**](Request.md) |  | 

### Return type

[**ProviderValidationResponse**](ProviderValidationResponse.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

