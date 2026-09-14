# \ModelDeploymentPoolsAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddPoolMemberV1ModelPoolsPoolIdMembersPost**](ModelDeploymentPoolsAPI.md#AddPoolMemberV1ModelPoolsPoolIdMembersPost) | **Post** /v1/model-pools/{pool_id}/members | Add Pool Member
[**CreateModelPoolV1ModelPoolsPost**](ModelDeploymentPoolsAPI.md#CreateModelPoolV1ModelPoolsPost) | **Post** /v1/model-pools | Create Model Pool
[**DeleteModelPoolV1ModelPoolsPoolIdDelete**](ModelDeploymentPoolsAPI.md#DeleteModelPoolV1ModelPoolsPoolIdDelete) | **Delete** /v1/model-pools/{pool_id} | Delete Model Pool
[**GetModelPoolV1ModelPoolsPoolIdGet**](ModelDeploymentPoolsAPI.md#GetModelPoolV1ModelPoolsPoolIdGet) | **Get** /v1/model-pools/{pool_id} | Get Model Pool
[**ListModelPoolsV1ModelPoolsGet**](ModelDeploymentPoolsAPI.md#ListModelPoolsV1ModelPoolsGet) | **Get** /v1/model-pools | List Model Pools
[**ListPoolMembersV1ModelPoolsPoolIdMembersGet**](ModelDeploymentPoolsAPI.md#ListPoolMembersV1ModelPoolsPoolIdMembersGet) | **Get** /v1/model-pools/{pool_id}/members | List Pool Members
[**RemovePoolMemberV1ModelPoolsPoolIdMembersMemberIdDelete**](ModelDeploymentPoolsAPI.md#RemovePoolMemberV1ModelPoolsPoolIdMembersMemberIdDelete) | **Delete** /v1/model-pools/{pool_id}/members/{member_id} | Remove Pool Member
[**UpdateModelPoolV1ModelPoolsPoolIdPut**](ModelDeploymentPoolsAPI.md#UpdateModelPoolV1ModelPoolsPoolIdPut) | **Put** /v1/model-pools/{pool_id} | Update Model Pool
[**UpdatePoolMemberV1ModelPoolsPoolIdMembersMemberIdPut**](ModelDeploymentPoolsAPI.md#UpdatePoolMemberV1ModelPoolsPoolIdMembersMemberIdPut) | **Put** /v1/model-pools/{pool_id}/members/{member_id} | Update Pool Member



## AddPoolMemberV1ModelPoolsPoolIdMembersPost

> PoolMember AddPoolMemberV1ModelPoolsPoolIdMembersPost(ctx, poolId).PoolMemberCreate(poolMemberCreate).Execute()

Add Pool Member



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
	poolId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 
	poolMemberCreate := *openapiclient.NewPoolMemberCreate("ModelDeploymentId_example") // PoolMemberCreate | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ModelDeploymentPoolsAPI.AddPoolMemberV1ModelPoolsPoolIdMembersPost(context.Background(), poolId).PoolMemberCreate(poolMemberCreate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ModelDeploymentPoolsAPI.AddPoolMemberV1ModelPoolsPoolIdMembersPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AddPoolMemberV1ModelPoolsPoolIdMembersPost`: PoolMember
	fmt.Fprintf(os.Stdout, "Response from `ModelDeploymentPoolsAPI.AddPoolMemberV1ModelPoolsPoolIdMembersPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**poolId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiAddPoolMemberV1ModelPoolsPoolIdMembersPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **poolMemberCreate** | [**PoolMemberCreate**](PoolMemberCreate.md) |  | 

### Return type

[**PoolMember**](PoolMember.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateModelPoolV1ModelPoolsPost

> ModelDeploymentPool CreateModelPoolV1ModelPoolsPost(ctx).ModelDeploymentPoolCreate(modelDeploymentPoolCreate).Execute()

Create Model Pool



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
	modelDeploymentPoolCreate := *openapiclient.NewModelDeploymentPoolCreate("Name_example") // ModelDeploymentPoolCreate | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ModelDeploymentPoolsAPI.CreateModelPoolV1ModelPoolsPost(context.Background()).ModelDeploymentPoolCreate(modelDeploymentPoolCreate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ModelDeploymentPoolsAPI.CreateModelPoolV1ModelPoolsPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateModelPoolV1ModelPoolsPost`: ModelDeploymentPool
	fmt.Fprintf(os.Stdout, "Response from `ModelDeploymentPoolsAPI.CreateModelPoolV1ModelPoolsPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateModelPoolV1ModelPoolsPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **modelDeploymentPoolCreate** | [**ModelDeploymentPoolCreate**](ModelDeploymentPoolCreate.md) |  | 

### Return type

[**ModelDeploymentPool**](ModelDeploymentPool.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteModelPoolV1ModelPoolsPoolIdDelete

> DeleteModelPoolV1ModelPoolsPoolIdDelete(ctx, poolId).Execute()

Delete Model Pool



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
	poolId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ModelDeploymentPoolsAPI.DeleteModelPoolV1ModelPoolsPoolIdDelete(context.Background(), poolId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ModelDeploymentPoolsAPI.DeleteModelPoolV1ModelPoolsPoolIdDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**poolId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteModelPoolV1ModelPoolsPoolIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetModelPoolV1ModelPoolsPoolIdGet

> ModelDeploymentPool GetModelPoolV1ModelPoolsPoolIdGet(ctx, poolId).Execute()

Get Model Pool



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
	poolId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ModelDeploymentPoolsAPI.GetModelPoolV1ModelPoolsPoolIdGet(context.Background(), poolId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ModelDeploymentPoolsAPI.GetModelPoolV1ModelPoolsPoolIdGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetModelPoolV1ModelPoolsPoolIdGet`: ModelDeploymentPool
	fmt.Fprintf(os.Stdout, "Response from `ModelDeploymentPoolsAPI.GetModelPoolV1ModelPoolsPoolIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**poolId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetModelPoolV1ModelPoolsPoolIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ModelDeploymentPool**](ModelDeploymentPool.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListModelPoolsV1ModelPoolsGet

> PagedResponseModelModelDeploymentPool ListModelPoolsV1ModelPoolsGet(ctx).Page(page).Size(size).Sort(sort).Filter(filter).Execute()

List Model Pools



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
	page := int32(56) // int32 |  (optional) (default to 1)
	size := int32(56) // int32 |  (optional) (default to 10)
	sort := "sort_example" // string |  (optional) (default to "id")
	filter := "filter_example" // string |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ModelDeploymentPoolsAPI.ListModelPoolsV1ModelPoolsGet(context.Background()).Page(page).Size(size).Sort(sort).Filter(filter).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ModelDeploymentPoolsAPI.ListModelPoolsV1ModelPoolsGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListModelPoolsV1ModelPoolsGet`: PagedResponseModelModelDeploymentPool
	fmt.Fprintf(os.Stdout, "Response from `ModelDeploymentPoolsAPI.ListModelPoolsV1ModelPoolsGet`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListModelPoolsV1ModelPoolsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int32** |  | [default to 1]
 **size** | **int32** |  | [default to 10]
 **sort** | **string** |  | [default to &quot;id&quot;]
 **filter** | **string** |  | 

### Return type

[**PagedResponseModelModelDeploymentPool**](PagedResponseModelModelDeploymentPool.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListPoolMembersV1ModelPoolsPoolIdMembersGet

> []PoolMember ListPoolMembersV1ModelPoolsPoolIdMembersGet(ctx, poolId).Execute()

List Pool Members



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
	poolId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ModelDeploymentPoolsAPI.ListPoolMembersV1ModelPoolsPoolIdMembersGet(context.Background(), poolId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ModelDeploymentPoolsAPI.ListPoolMembersV1ModelPoolsPoolIdMembersGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListPoolMembersV1ModelPoolsPoolIdMembersGet`: []PoolMember
	fmt.Fprintf(os.Stdout, "Response from `ModelDeploymentPoolsAPI.ListPoolMembersV1ModelPoolsPoolIdMembersGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**poolId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiListPoolMembersV1ModelPoolsPoolIdMembersGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]PoolMember**](PoolMember.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RemovePoolMemberV1ModelPoolsPoolIdMembersMemberIdDelete

> RemovePoolMemberV1ModelPoolsPoolIdMembersMemberIdDelete(ctx, poolId, memberId).Execute()

Remove Pool Member



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
	poolId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 
	memberId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ModelDeploymentPoolsAPI.RemovePoolMemberV1ModelPoolsPoolIdMembersMemberIdDelete(context.Background(), poolId, memberId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ModelDeploymentPoolsAPI.RemovePoolMemberV1ModelPoolsPoolIdMembersMemberIdDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**poolId** | **string** |  | 
**memberId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiRemovePoolMemberV1ModelPoolsPoolIdMembersMemberIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

 (empty response body)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateModelPoolV1ModelPoolsPoolIdPut

> ModelDeploymentPool UpdateModelPoolV1ModelPoolsPoolIdPut(ctx, poolId).ModelDeploymentPoolUpdate(modelDeploymentPoolUpdate).Execute()

Update Model Pool



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
	poolId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 
	modelDeploymentPoolUpdate := *openapiclient.NewModelDeploymentPoolUpdate() // ModelDeploymentPoolUpdate | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ModelDeploymentPoolsAPI.UpdateModelPoolV1ModelPoolsPoolIdPut(context.Background(), poolId).ModelDeploymentPoolUpdate(modelDeploymentPoolUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ModelDeploymentPoolsAPI.UpdateModelPoolV1ModelPoolsPoolIdPut``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateModelPoolV1ModelPoolsPoolIdPut`: ModelDeploymentPool
	fmt.Fprintf(os.Stdout, "Response from `ModelDeploymentPoolsAPI.UpdateModelPoolV1ModelPoolsPoolIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**poolId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateModelPoolV1ModelPoolsPoolIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **modelDeploymentPoolUpdate** | [**ModelDeploymentPoolUpdate**](ModelDeploymentPoolUpdate.md) |  | 

### Return type

[**ModelDeploymentPool**](ModelDeploymentPool.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdatePoolMemberV1ModelPoolsPoolIdMembersMemberIdPut

> PoolMember UpdatePoolMemberV1ModelPoolsPoolIdMembersMemberIdPut(ctx, poolId, memberId).PoolMemberUpdate(poolMemberUpdate).Execute()

Update Pool Member



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
	poolId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 
	memberId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 
	poolMemberUpdate := *openapiclient.NewPoolMemberUpdate() // PoolMemberUpdate | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ModelDeploymentPoolsAPI.UpdatePoolMemberV1ModelPoolsPoolIdMembersMemberIdPut(context.Background(), poolId, memberId).PoolMemberUpdate(poolMemberUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ModelDeploymentPoolsAPI.UpdatePoolMemberV1ModelPoolsPoolIdMembersMemberIdPut``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdatePoolMemberV1ModelPoolsPoolIdMembersMemberIdPut`: PoolMember
	fmt.Fprintf(os.Stdout, "Response from `ModelDeploymentPoolsAPI.UpdatePoolMemberV1ModelPoolsPoolIdMembersMemberIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**poolId** | **string** |  | 
**memberId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdatePoolMemberV1ModelPoolsPoolIdMembersMemberIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **poolMemberUpdate** | [**PoolMemberUpdate**](PoolMemberUpdate.md) |  | 

### Return type

[**PoolMember**](PoolMember.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

