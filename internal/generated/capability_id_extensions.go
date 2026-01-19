// Copyright (c) Trifork
//
// This file adds missing methods to CapabilityId and CapabilityId1 types.
// The OpenAPI generator doesn't generate GetActualInstanceValue() for simple
// anyOf types that just wrap a string, which causes path parameter serialization
// to fail. This file can be safely kept across regenerations as it only adds methods.

package api

// GetActualInstance returns the actual instance of the CapabilityId.
func (obj *CapabilityId) GetActualInstance() interface{} {
	if obj == nil {
		return nil
	}
	if obj.String != nil {
		return obj.String
	}
	return nil
}

// GetActualInstanceValue returns the actual instance value of the CapabilityId.
func (obj CapabilityId) GetActualInstanceValue() interface{} {
	if obj.String != nil {
		return *obj.String
	}
	return nil
}

// GetActualInstance returns the actual instance of the CapabilityId1.
func (obj *CapabilityId1) GetActualInstance() interface{} {
	if obj == nil {
		return nil
	}
	if obj.String != nil {
		return obj.String
	}
	return nil
}

// GetActualInstanceValue returns the actual instance value of the CapabilityId1.
func (obj CapabilityId1) GetActualInstanceValue() interface{} {
	if obj.String != nil {
		return *obj.String
	}
	return nil
}
