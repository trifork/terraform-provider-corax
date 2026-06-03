# \ComplianceAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ApplyControlsV1ComplianceCapabilitiesCapabilityIdPoliciesPolicyIdPut**](ComplianceAPI.md#ApplyControlsV1ComplianceCapabilitiesCapabilityIdPoliciesPolicyIdPut) | **Put** /v1/compliance/capabilities/{capability_id}/policies/{policy_id} | Apply Controls
[**ApplyPolicyToAllV1CompliancePoliciesPolicyIdApplicationsPost**](ComplianceAPI.md#ApplyPolicyToAllV1CompliancePoliciesPolicyIdApplicationsPost) | **Post** /v1/compliance/policies/{policy_id}/applications | Apply Policy To All
[**CreatePolicyV1CompliancePoliciesPost**](ComplianceAPI.md#CreatePolicyV1CompliancePoliciesPost) | **Post** /v1/compliance/policies | Create Policy
[**DeletePolicyV1CompliancePoliciesPolicyIdDelete**](ComplianceAPI.md#DeletePolicyV1CompliancePoliciesPolicyIdDelete) | **Delete** /v1/compliance/policies/{policy_id} | Delete Policy
[**DryRunGuardrailsV1ComplianceGuardrailScansPost**](ComplianceAPI.md#DryRunGuardrailsV1ComplianceGuardrailScansPost) | **Post** /v1/compliance/guardrail-scans | Dry Run Guardrails
[**DuplicatePolicyV1CompliancePoliciesPolicyIdCopiesPost**](ComplianceAPI.md#DuplicatePolicyV1CompliancePoliciesPolicyIdCopiesPost) | **Post** /v1/compliance/policies/{policy_id}/copies | Duplicate Policy
[**GetApplicablePoliciesV1ComplianceCapabilitiesCapabilityIdPoliciesGet**](ComplianceAPI.md#GetApplicablePoliciesV1ComplianceCapabilitiesCapabilityIdPoliciesGet) | **Get** /v1/compliance/capabilities/{capability_id}/policies | Get Applicable Policies
[**GetAssetDetailV1ComplianceAssetsCapabilityIdGet**](ComplianceAPI.md#GetAssetDetailV1ComplianceAssetsCapabilityIdGet) | **Get** /v1/compliance/assets/{capability_id} | Get Asset Detail
[**GetComplianceBadgeV1ComplianceAssetsCapabilityIdBadgeGet**](ComplianceAPI.md#GetComplianceBadgeV1ComplianceAssetsCapabilityIdBadgeGet) | **Get** /v1/compliance/assets/{capability_id}/badge | Get Compliance Badge
[**GetGuardrailsV1ComplianceGuardrailsGet**](ComplianceAPI.md#GetGuardrailsV1ComplianceGuardrailsGet) | **Get** /v1/compliance/guardrails | Get Guardrails
[**GetPolicyImpactV1CompliancePoliciesPolicyIdImpactGet**](ComplianceAPI.md#GetPolicyImpactV1CompliancePoliciesPolicyIdImpactGet) | **Get** /v1/compliance/policies/{policy_id}/impact | Get Policy Impact
[**GetPolicyV1CompliancePoliciesPolicyIdGet**](ComplianceAPI.md#GetPolicyV1CompliancePoliciesPolicyIdGet) | **Get** /v1/compliance/policies/{policy_id} | Get Policy
[**ListAssetsV1ComplianceAssetsGet**](ComplianceAPI.md#ListAssetsV1ComplianceAssetsGet) | **Get** /v1/compliance/assets | List Assets
[**ListGuardrailEventsV1ComplianceGuardrailEventsGet**](ComplianceAPI.md#ListGuardrailEventsV1ComplianceGuardrailEventsGet) | **Get** /v1/compliance/guardrail-events | List Guardrail Events
[**ListPoliciesV1CompliancePoliciesGet**](ComplianceAPI.md#ListPoliciesV1CompliancePoliciesGet) | **Get** /v1/compliance/policies | List Policies
[**RemoveControlsV1ComplianceCapabilitiesCapabilityIdPoliciesPolicyIdDelete**](ComplianceAPI.md#RemoveControlsV1ComplianceCapabilitiesCapabilityIdPoliciesPolicyIdDelete) | **Delete** /v1/compliance/capabilities/{capability_id}/policies/{policy_id} | Remove Controls
[**UpdatePolicyV1CompliancePoliciesPolicyIdPatch**](ComplianceAPI.md#UpdatePolicyV1CompliancePoliciesPolicyIdPatch) | **Patch** /v1/compliance/policies/{policy_id} | Update Policy



## ApplyControlsV1ComplianceCapabilitiesCapabilityIdPoliciesPolicyIdPut

> ApplyControlsResponse ApplyControlsV1ComplianceCapabilitiesCapabilityIdPoliciesPolicyIdPut(ctx, capabilityId, policyId).Execute()

Apply Controls



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
	capabilityId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 
	policyId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComplianceAPI.ApplyControlsV1ComplianceCapabilitiesCapabilityIdPoliciesPolicyIdPut(context.Background(), capabilityId, policyId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComplianceAPI.ApplyControlsV1ComplianceCapabilitiesCapabilityIdPoliciesPolicyIdPut``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ApplyControlsV1ComplianceCapabilitiesCapabilityIdPoliciesPolicyIdPut`: ApplyControlsResponse
	fmt.Fprintf(os.Stdout, "Response from `ComplianceAPI.ApplyControlsV1ComplianceCapabilitiesCapabilityIdPoliciesPolicyIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**capabilityId** | **string** |  | 
**policyId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApplyControlsV1ComplianceCapabilitiesCapabilityIdPoliciesPolicyIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**ApplyControlsResponse**](ApplyControlsResponse.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApplyPolicyToAllV1CompliancePoliciesPolicyIdApplicationsPost

> interface{} ApplyPolicyToAllV1CompliancePoliciesPolicyIdApplicationsPost(ctx, policyId).Execute()

Apply Policy To All



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
	policyId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComplianceAPI.ApplyPolicyToAllV1CompliancePoliciesPolicyIdApplicationsPost(context.Background(), policyId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComplianceAPI.ApplyPolicyToAllV1CompliancePoliciesPolicyIdApplicationsPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ApplyPolicyToAllV1CompliancePoliciesPolicyIdApplicationsPost`: interface{}
	fmt.Fprintf(os.Stdout, "Response from `ComplianceAPI.ApplyPolicyToAllV1CompliancePoliciesPolicyIdApplicationsPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**policyId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApplyPolicyToAllV1CompliancePoliciesPolicyIdApplicationsPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

**interface{}**

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreatePolicyV1CompliancePoliciesPost

> PolicyResponse CreatePolicyV1CompliancePoliciesPost(ctx).PolicyCreate(policyCreate).Execute()

Create Policy

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
	policyCreate := *openapiclient.NewPolicyCreate("Name_example", openapiclient.PolicyScopeType("global"), *openapiclient.NewPolicyControlsInput()) // PolicyCreate | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComplianceAPI.CreatePolicyV1CompliancePoliciesPost(context.Background()).PolicyCreate(policyCreate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComplianceAPI.CreatePolicyV1CompliancePoliciesPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreatePolicyV1CompliancePoliciesPost`: PolicyResponse
	fmt.Fprintf(os.Stdout, "Response from `ComplianceAPI.CreatePolicyV1CompliancePoliciesPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreatePolicyV1CompliancePoliciesPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **policyCreate** | [**PolicyCreate**](PolicyCreate.md) |  | 

### Return type

[**PolicyResponse**](PolicyResponse.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeletePolicyV1CompliancePoliciesPolicyIdDelete

> DeletePolicyV1CompliancePoliciesPolicyIdDelete(ctx, policyId).Execute()

Delete Policy

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
	policyId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ComplianceAPI.DeletePolicyV1CompliancePoliciesPolicyIdDelete(context.Background(), policyId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComplianceAPI.DeletePolicyV1CompliancePoliciesPolicyIdDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**policyId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeletePolicyV1CompliancePoliciesPolicyIdDeleteRequest struct via the builder pattern


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


## DryRunGuardrailsV1ComplianceGuardrailScansPost

> DryRunResponse DryRunGuardrailsV1ComplianceGuardrailScansPost(ctx).DryRunRequest(dryRunRequest).Execute()

Dry Run Guardrails



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
	dryRunRequest := *openapiclient.NewDryRunRequest("Text_example") // DryRunRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComplianceAPI.DryRunGuardrailsV1ComplianceGuardrailScansPost(context.Background()).DryRunRequest(dryRunRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComplianceAPI.DryRunGuardrailsV1ComplianceGuardrailScansPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DryRunGuardrailsV1ComplianceGuardrailScansPost`: DryRunResponse
	fmt.Fprintf(os.Stdout, "Response from `ComplianceAPI.DryRunGuardrailsV1ComplianceGuardrailScansPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDryRunGuardrailsV1ComplianceGuardrailScansPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **dryRunRequest** | [**DryRunRequest**](DryRunRequest.md) |  | 

### Return type

[**DryRunResponse**](DryRunResponse.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DuplicatePolicyV1CompliancePoliciesPolicyIdCopiesPost

> PolicyResponse DuplicatePolicyV1CompliancePoliciesPolicyIdCopiesPost(ctx, policyId).Execute()

Duplicate Policy



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
	policyId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComplianceAPI.DuplicatePolicyV1CompliancePoliciesPolicyIdCopiesPost(context.Background(), policyId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComplianceAPI.DuplicatePolicyV1CompliancePoliciesPolicyIdCopiesPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DuplicatePolicyV1CompliancePoliciesPolicyIdCopiesPost`: PolicyResponse
	fmt.Fprintf(os.Stdout, "Response from `ComplianceAPI.DuplicatePolicyV1CompliancePoliciesPolicyIdCopiesPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**policyId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiDuplicatePolicyV1CompliancePoliciesPolicyIdCopiesPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**PolicyResponse**](PolicyResponse.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetApplicablePoliciesV1ComplianceCapabilitiesCapabilityIdPoliciesGet

> []ApplicablePolicyResponse GetApplicablePoliciesV1ComplianceCapabilitiesCapabilityIdPoliciesGet(ctx, capabilityId).Execute()

Get Applicable Policies



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
	capabilityId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComplianceAPI.GetApplicablePoliciesV1ComplianceCapabilitiesCapabilityIdPoliciesGet(context.Background(), capabilityId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComplianceAPI.GetApplicablePoliciesV1ComplianceCapabilitiesCapabilityIdPoliciesGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetApplicablePoliciesV1ComplianceCapabilitiesCapabilityIdPoliciesGet`: []ApplicablePolicyResponse
	fmt.Fprintf(os.Stdout, "Response from `ComplianceAPI.GetApplicablePoliciesV1ComplianceCapabilitiesCapabilityIdPoliciesGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**capabilityId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetApplicablePoliciesV1ComplianceCapabilitiesCapabilityIdPoliciesGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]ApplicablePolicyResponse**](ApplicablePolicyResponse.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetAssetDetailV1ComplianceAssetsCapabilityIdGet

> AssetResponse GetAssetDetailV1ComplianceAssetsCapabilityIdGet(ctx, capabilityId).Execute()

Get Asset Detail

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
	capabilityId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComplianceAPI.GetAssetDetailV1ComplianceAssetsCapabilityIdGet(context.Background(), capabilityId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComplianceAPI.GetAssetDetailV1ComplianceAssetsCapabilityIdGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetAssetDetailV1ComplianceAssetsCapabilityIdGet`: AssetResponse
	fmt.Fprintf(os.Stdout, "Response from `ComplianceAPI.GetAssetDetailV1ComplianceAssetsCapabilityIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**capabilityId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetAssetDetailV1ComplianceAssetsCapabilityIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**AssetResponse**](AssetResponse.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetComplianceBadgeV1ComplianceAssetsCapabilityIdBadgeGet

> BadgeResponse GetComplianceBadgeV1ComplianceAssetsCapabilityIdBadgeGet(ctx, capabilityId).Execute()

Get Compliance Badge



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
	capabilityId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComplianceAPI.GetComplianceBadgeV1ComplianceAssetsCapabilityIdBadgeGet(context.Background(), capabilityId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComplianceAPI.GetComplianceBadgeV1ComplianceAssetsCapabilityIdBadgeGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetComplianceBadgeV1ComplianceAssetsCapabilityIdBadgeGet`: BadgeResponse
	fmt.Fprintf(os.Stdout, "Response from `ComplianceAPI.GetComplianceBadgeV1ComplianceAssetsCapabilityIdBadgeGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**capabilityId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetComplianceBadgeV1ComplianceAssetsCapabilityIdBadgeGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**BadgeResponse**](BadgeResponse.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetGuardrailsV1ComplianceGuardrailsGet

> []GuardrailRow GetGuardrailsV1ComplianceGuardrailsGet(ctx).ProjectId(projectId).SortBy(sortBy).SortOrder(sortOrder).Execute()

Get Guardrails

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
	projectId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string |  (optional)
	sortBy := "sortBy_example" // string |  (optional) (default to "capability_name")
	sortOrder := "sortOrder_example" // string |  (optional) (default to "asc")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComplianceAPI.GetGuardrailsV1ComplianceGuardrailsGet(context.Background()).ProjectId(projectId).SortBy(sortBy).SortOrder(sortOrder).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComplianceAPI.GetGuardrailsV1ComplianceGuardrailsGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetGuardrailsV1ComplianceGuardrailsGet`: []GuardrailRow
	fmt.Fprintf(os.Stdout, "Response from `ComplianceAPI.GetGuardrailsV1ComplianceGuardrailsGet`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetGuardrailsV1ComplianceGuardrailsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **projectId** | **string** |  | 
 **sortBy** | **string** |  | [default to &quot;capability_name&quot;]
 **sortOrder** | **string** |  | [default to &quot;asc&quot;]

### Return type

[**[]GuardrailRow**](GuardrailRow.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetPolicyImpactV1CompliancePoliciesPolicyIdImpactGet

> PolicyImpactResponse GetPolicyImpactV1CompliancePoliciesPolicyIdImpactGet(ctx, policyId).Execute()

Get Policy Impact



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
	policyId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComplianceAPI.GetPolicyImpactV1CompliancePoliciesPolicyIdImpactGet(context.Background(), policyId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComplianceAPI.GetPolicyImpactV1CompliancePoliciesPolicyIdImpactGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetPolicyImpactV1CompliancePoliciesPolicyIdImpactGet`: PolicyImpactResponse
	fmt.Fprintf(os.Stdout, "Response from `ComplianceAPI.GetPolicyImpactV1CompliancePoliciesPolicyIdImpactGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**policyId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetPolicyImpactV1CompliancePoliciesPolicyIdImpactGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**PolicyImpactResponse**](PolicyImpactResponse.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetPolicyV1CompliancePoliciesPolicyIdGet

> PolicyResponse GetPolicyV1CompliancePoliciesPolicyIdGet(ctx, policyId).Execute()

Get Policy

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
	policyId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComplianceAPI.GetPolicyV1CompliancePoliciesPolicyIdGet(context.Background(), policyId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComplianceAPI.GetPolicyV1CompliancePoliciesPolicyIdGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetPolicyV1CompliancePoliciesPolicyIdGet`: PolicyResponse
	fmt.Fprintf(os.Stdout, "Response from `ComplianceAPI.GetPolicyV1CompliancePoliciesPolicyIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**policyId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetPolicyV1CompliancePoliciesPolicyIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**PolicyResponse**](PolicyResponse.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListAssetsV1ComplianceAssetsGet

> []AssetResponse ListAssetsV1ComplianceAssetsGet(ctx).ProjectId(projectId).Status(status).Execute()

List Assets

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
	projectId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string |  (optional)
	status := openapiclient.ComplianceStatus("protected") // ComplianceStatus |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComplianceAPI.ListAssetsV1ComplianceAssetsGet(context.Background()).ProjectId(projectId).Status(status).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComplianceAPI.ListAssetsV1ComplianceAssetsGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListAssetsV1ComplianceAssetsGet`: []AssetResponse
	fmt.Fprintf(os.Stdout, "Response from `ComplianceAPI.ListAssetsV1ComplianceAssetsGet`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListAssetsV1ComplianceAssetsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **projectId** | **string** |  | 
 **status** | [**ComplianceStatus**](ComplianceStatus.md) |  | 

### Return type

[**[]AssetResponse**](AssetResponse.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListGuardrailEventsV1ComplianceGuardrailEventsGet

> GuardrailEventListResponse ListGuardrailEventsV1ComplianceGuardrailEventsGet(ctx).CapabilityId(capabilityId).PolicyId(policyId).ControlType(controlType).Direction(direction).FromDate(fromDate).ToDate(toDate).Limit(limit).Offset(offset).Execute()

List Guardrail Events



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID/api"
)

func main() {
	capabilityId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string |  (optional)
	policyId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string |  (optional)
	controlType := openapiclient.ScanCategory("content_filter") // ScanCategory |  (optional)
	direction := openapiclient.ScanDirection("input") // ScanDirection |  (optional)
	fromDate := time.Now() // time.Time |  (optional)
	toDate := time.Now() // time.Time |  (optional)
	limit := int32(56) // int32 |  (optional) (default to 50)
	offset := int32(56) // int32 |  (optional) (default to 0)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComplianceAPI.ListGuardrailEventsV1ComplianceGuardrailEventsGet(context.Background()).CapabilityId(capabilityId).PolicyId(policyId).ControlType(controlType).Direction(direction).FromDate(fromDate).ToDate(toDate).Limit(limit).Offset(offset).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComplianceAPI.ListGuardrailEventsV1ComplianceGuardrailEventsGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListGuardrailEventsV1ComplianceGuardrailEventsGet`: GuardrailEventListResponse
	fmt.Fprintf(os.Stdout, "Response from `ComplianceAPI.ListGuardrailEventsV1ComplianceGuardrailEventsGet`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListGuardrailEventsV1ComplianceGuardrailEventsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **capabilityId** | **string** |  | 
 **policyId** | **string** |  | 
 **controlType** | [**ScanCategory**](ScanCategory.md) |  | 
 **direction** | [**ScanDirection**](ScanDirection.md) |  | 
 **fromDate** | **time.Time** |  | 
 **toDate** | **time.Time** |  | 
 **limit** | **int32** |  | [default to 50]
 **offset** | **int32** |  | [default to 0]

### Return type

[**GuardrailEventListResponse**](GuardrailEventListResponse.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListPoliciesV1CompliancePoliciesGet

> []PolicyResponse ListPoliciesV1CompliancePoliciesGet(ctx).Execute()

List Policies

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
	resp, r, err := apiClient.ComplianceAPI.ListPoliciesV1CompliancePoliciesGet(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComplianceAPI.ListPoliciesV1CompliancePoliciesGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListPoliciesV1CompliancePoliciesGet`: []PolicyResponse
	fmt.Fprintf(os.Stdout, "Response from `ComplianceAPI.ListPoliciesV1CompliancePoliciesGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiListPoliciesV1CompliancePoliciesGetRequest struct via the builder pattern


### Return type

[**[]PolicyResponse**](PolicyResponse.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RemoveControlsV1ComplianceCapabilitiesCapabilityIdPoliciesPolicyIdDelete

> RemoveControlsResponse RemoveControlsV1ComplianceCapabilitiesCapabilityIdPoliciesPolicyIdDelete(ctx, capabilityId, policyId).Execute()

Remove Controls



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
	capabilityId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 
	policyId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComplianceAPI.RemoveControlsV1ComplianceCapabilitiesCapabilityIdPoliciesPolicyIdDelete(context.Background(), capabilityId, policyId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComplianceAPI.RemoveControlsV1ComplianceCapabilitiesCapabilityIdPoliciesPolicyIdDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `RemoveControlsV1ComplianceCapabilitiesCapabilityIdPoliciesPolicyIdDelete`: RemoveControlsResponse
	fmt.Fprintf(os.Stdout, "Response from `ComplianceAPI.RemoveControlsV1ComplianceCapabilitiesCapabilityIdPoliciesPolicyIdDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**capabilityId** | **string** |  | 
**policyId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiRemoveControlsV1ComplianceCapabilitiesCapabilityIdPoliciesPolicyIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**RemoveControlsResponse**](RemoveControlsResponse.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdatePolicyV1CompliancePoliciesPolicyIdPatch

> PolicyResponse UpdatePolicyV1CompliancePoliciesPolicyIdPatch(ctx, policyId).PolicyUpdate(policyUpdate).Execute()

Update Policy

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
	policyId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 
	policyUpdate := *openapiclient.NewPolicyUpdate() // PolicyUpdate | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ComplianceAPI.UpdatePolicyV1CompliancePoliciesPolicyIdPatch(context.Background(), policyId).PolicyUpdate(policyUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ComplianceAPI.UpdatePolicyV1CompliancePoliciesPolicyIdPatch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdatePolicyV1CompliancePoliciesPolicyIdPatch`: PolicyResponse
	fmt.Fprintf(os.Stdout, "Response from `ComplianceAPI.UpdatePolicyV1CompliancePoliciesPolicyIdPatch`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**policyId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdatePolicyV1CompliancePoliciesPolicyIdPatchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **policyUpdate** | [**PolicyUpdate**](PolicyUpdate.md) |  | 

### Return type

[**PolicyResponse**](PolicyResponse.md)

### Authorization

[APIKeyHeader](../README.md#APIKeyHeader), [HTTPBearer](../README.md#HTTPBearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

