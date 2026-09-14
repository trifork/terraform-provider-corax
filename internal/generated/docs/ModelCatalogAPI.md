# \ModelCatalogAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetModelCatalogEntry**](ModelCatalogAPI.md#GetModelCatalogEntry) | **Get** /v1/model-catalog/{catalog_id} | Fetch a single catalog entry by stable catalog ID.
[**ListModelCatalogEntries**](ModelCatalogAPI.md#ListModelCatalogEntries) | **Get** /v1/model-catalog | List the entire normalized model catalog.
[**ListModelCatalogProviders**](ModelCatalogAPI.md#ListModelCatalogProviders) | **Get** /v1/model-catalog/providers | List model providers available in the catalog.



## GetModelCatalogEntry

> ModelCatalogEntryResponse GetModelCatalogEntry(ctx, catalogId).Execute()

Fetch a single catalog entry by stable catalog ID.



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
	catalogId := "catalogId_example" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ModelCatalogAPI.GetModelCatalogEntry(context.Background(), catalogId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ModelCatalogAPI.GetModelCatalogEntry``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetModelCatalogEntry`: ModelCatalogEntryResponse
	fmt.Fprintf(os.Stdout, "Response from `ModelCatalogAPI.GetModelCatalogEntry`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**catalogId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetModelCatalogEntryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ModelCatalogEntryResponse**](ModelCatalogEntryResponse.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListModelCatalogEntries

> ModelCatalogListResponse ListModelCatalogEntries(ctx).Provider(provider).Execute()

List the entire normalized model catalog.



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
	provider := "provider_example" // string |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ModelCatalogAPI.ListModelCatalogEntries(context.Background()).Provider(provider).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ModelCatalogAPI.ListModelCatalogEntries``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListModelCatalogEntries`: ModelCatalogListResponse
	fmt.Fprintf(os.Stdout, "Response from `ModelCatalogAPI.ListModelCatalogEntries`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListModelCatalogEntriesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **provider** | **string** |  | 

### Return type

[**ModelCatalogListResponse**](ModelCatalogListResponse.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListModelCatalogProviders

> ProviderListResponse ListModelCatalogProviders(ctx).Execute()

List model providers available in the catalog.



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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ModelCatalogAPI.ListModelCatalogProviders(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ModelCatalogAPI.ListModelCatalogProviders``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListModelCatalogProviders`: ProviderListResponse
	fmt.Fprintf(os.Stdout, "Response from `ModelCatalogAPI.ListModelCatalogProviders`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListModelCatalogProvidersRequest struct via the builder pattern


### Return type

[**ProviderListResponse**](ProviderListResponse.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

