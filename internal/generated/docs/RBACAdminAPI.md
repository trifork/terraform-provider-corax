# \RBACAdminAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateRoleV1AdminRolesPost**](RBACAdminAPI.md#CreateRoleV1AdminRolesPost) | **Post** /v1/admin/roles | Create Role
[**DeleteRoleV1AdminRolesRoleIdDelete**](RBACAdminAPI.md#DeleteRoleV1AdminRolesRoleIdDelete) | **Delete** /v1/admin/roles/{role_id} | Delete Role
[**ListPermissionsV1AdminPermissionsGet**](RBACAdminAPI.md#ListPermissionsV1AdminPermissionsGet) | **Get** /v1/admin/permissions | List Permissions
[**ListRolesV1AdminRolesGet**](RBACAdminAPI.md#ListRolesV1AdminRolesGet) | **Get** /v1/admin/roles | List Roles
[**ListUsersV1AdminUsersGet**](RBACAdminAPI.md#ListUsersV1AdminUsersGet) | **Get** /v1/admin/users | List Users
[**UpdateRoleV1AdminRolesRoleIdPut**](RBACAdminAPI.md#UpdateRoleV1AdminRolesRoleIdPut) | **Put** /v1/admin/roles/{role_id} | Update Role
[**UpdateUserV1AdminUsersUserIdPatch**](RBACAdminAPI.md#UpdateUserV1AdminUsersUserIdPatch) | **Patch** /v1/admin/users/{user_id} | Update User



## CreateRoleV1AdminRolesPost

> Role CreateRoleV1AdminRolesPost(ctx).RoleCreate(roleCreate).Execute()

Create Role

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
	roleCreate := *openapiclient.NewRoleCreate("Name_example") // RoleCreate | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RBACAdminAPI.CreateRoleV1AdminRolesPost(context.Background()).RoleCreate(roleCreate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RBACAdminAPI.CreateRoleV1AdminRolesPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateRoleV1AdminRolesPost`: Role
	fmt.Fprintf(os.Stdout, "Response from `RBACAdminAPI.CreateRoleV1AdminRolesPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateRoleV1AdminRolesPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **roleCreate** | [**RoleCreate**](RoleCreate.md) |  | 

### Return type

[**Role**](Role.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteRoleV1AdminRolesRoleIdDelete

> DeleteRoleV1AdminRolesRoleIdDelete(ctx, roleId).Execute()

Delete Role

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
	roleId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.RBACAdminAPI.DeleteRoleV1AdminRolesRoleIdDelete(context.Background(), roleId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RBACAdminAPI.DeleteRoleV1AdminRolesRoleIdDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**roleId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteRoleV1AdminRolesRoleIdDeleteRequest struct via the builder pattern


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


## ListPermissionsV1AdminPermissionsGet

> []PermissionItem ListPermissionsV1AdminPermissionsGet(ctx).Execute()

List Permissions

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
	resp, r, err := apiClient.RBACAdminAPI.ListPermissionsV1AdminPermissionsGet(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RBACAdminAPI.ListPermissionsV1AdminPermissionsGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListPermissionsV1AdminPermissionsGet`: []PermissionItem
	fmt.Fprintf(os.Stdout, "Response from `RBACAdminAPI.ListPermissionsV1AdminPermissionsGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListPermissionsV1AdminPermissionsGetRequest struct via the builder pattern


### Return type

[**[]PermissionItem**](PermissionItem.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListRolesV1AdminRolesGet

> PagedResponseModelRole ListRolesV1AdminRolesGet(ctx).Page(page).Size(size).Sort(sort).Filter(filter).Execute()

List Roles

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
	resp, r, err := apiClient.RBACAdminAPI.ListRolesV1AdminRolesGet(context.Background()).Page(page).Size(size).Sort(sort).Filter(filter).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RBACAdminAPI.ListRolesV1AdminRolesGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListRolesV1AdminRolesGet`: PagedResponseModelRole
	fmt.Fprintf(os.Stdout, "Response from `RBACAdminAPI.ListRolesV1AdminRolesGet`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListRolesV1AdminRolesGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int32** |  | [default to 1]
 **size** | **int32** |  | [default to 10]
 **sort** | **string** |  | [default to &quot;id&quot;]
 **filter** | **string** |  | 

### Return type

[**PagedResponseModelRole**](PagedResponseModelRole.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListUsersV1AdminUsersGet

> PagedResponseModelUserRBAC ListUsersV1AdminUsersGet(ctx).Page(page).Size(size).Sort(sort).Filter(filter).Execute()

List Users

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
	resp, r, err := apiClient.RBACAdminAPI.ListUsersV1AdminUsersGet(context.Background()).Page(page).Size(size).Sort(sort).Filter(filter).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RBACAdminAPI.ListUsersV1AdminUsersGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListUsersV1AdminUsersGet`: PagedResponseModelUserRBAC
	fmt.Fprintf(os.Stdout, "Response from `RBACAdminAPI.ListUsersV1AdminUsersGet`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListUsersV1AdminUsersGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **int32** |  | [default to 1]
 **size** | **int32** |  | [default to 10]
 **sort** | **string** |  | [default to &quot;id&quot;]
 **filter** | **string** |  | 

### Return type

[**PagedResponseModelUserRBAC**](PagedResponseModelUserRBAC.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateRoleV1AdminRolesRoleIdPut

> Role UpdateRoleV1AdminRolesRoleIdPut(ctx, roleId).RoleUpdate(roleUpdate).Execute()

Update Role

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
	roleId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 
	roleUpdate := *openapiclient.NewRoleUpdate("Name_example") // RoleUpdate | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RBACAdminAPI.UpdateRoleV1AdminRolesRoleIdPut(context.Background(), roleId).RoleUpdate(roleUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RBACAdminAPI.UpdateRoleV1AdminRolesRoleIdPut``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateRoleV1AdminRolesRoleIdPut`: Role
	fmt.Fprintf(os.Stdout, "Response from `RBACAdminAPI.UpdateRoleV1AdminRolesRoleIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**roleId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateRoleV1AdminRolesRoleIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **roleUpdate** | [**RoleUpdate**](RoleUpdate.md) |  | 

### Return type

[**Role**](Role.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateUserV1AdminUsersUserIdPatch

> UserRBAC UpdateUserV1AdminUsersUserIdPatch(ctx, userId).UserUpdate(userUpdate).Execute()

Update User

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
	userId := "userId_example" // string | 
	userUpdate := *openapiclient.NewUserUpdate() // UserUpdate | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RBACAdminAPI.UpdateUserV1AdminUsersUserIdPatch(context.Background(), userId).UserUpdate(userUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RBACAdminAPI.UpdateUserV1AdminUsersUserIdPatch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateUserV1AdminUsersUserIdPatch`: UserRBAC
	fmt.Fprintf(os.Stdout, "Response from `RBACAdminAPI.UpdateUserV1AdminUsersUserIdPatch`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**userId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateUserV1AdminUsersUserIdPatchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **userUpdate** | [**UserUpdate**](UserUpdate.md) |  | 

### Return type

[**UserRBAC**](UserRBAC.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

