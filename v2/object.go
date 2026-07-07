package aptos

import (
	"context"
	"fmt"
	"strconv"
)

// ObjectCoreResourceType is the fully-qualified type of the resource every
// Aptos object carries.
const ObjectCoreResourceType = "0x1::object::ObjectCore"

// objectCoreTypeTag is the 0x1::object::ObjectCore struct type, used as the
// default type argument for 0x1::object::transfer.
var objectCoreTypeTag = TypeTag{Value: &StructTag{
	Address: AccountOne,
	Module:  "object",
	Name:    "ObjectCore",
}}

// ObjectCore mirrors the on-chain 0x1::object::ObjectCore resource, describing
// an object's ownership and transferability.
type ObjectCore struct {
	// Owner is the current owner of the object.
	Owner AccountAddress
	// AllowUngatedTransfer reports whether the object can be transferred with a
	// plain 0x1::object::transfer (false for soul-bound / gated objects).
	AllowUngatedTransfer bool
	// GuidCreationNum is the object's GUID creation counter.
	GuidCreationNum uint64
}

// GetObjectCore reads the 0x1::object::ObjectCore resource for an object
// address, returning its ownership and transfer configuration.
func GetObjectCore(ctx context.Context, client Client, object AccountAddress, opts ...ResourceOption) (*ObjectCore, error) {
	resource, err := client.AccountResource(ctx, object, ObjectCoreResourceType, opts...)
	if err != nil {
		return nil, err
	}
	return parseObjectCore(resource.Data)
}

// parseObjectCore converts a decoded ObjectCore resource's data map into a
// typed ObjectCore.
func parseObjectCore(data map[string]any) (*ObjectCore, error) {
	ownerStr, ok := data["owner"].(string)
	if !ok {
		return nil, fmt.Errorf("object core missing owner field")
	}
	owner, err := ParseAddress(ownerStr)
	if err != nil {
		return nil, fmt.Errorf("invalid object owner %q: %w", ownerStr, err)
	}

	core := &ObjectCore{Owner: owner}

	if v, ok := data["allow_ungated_transfer"].(bool); ok {
		core.AllowUngatedTransfer = v
	}

	// guid_creation_num is a u64 and is serialized as a JSON string.
	if guidStr, ok := data["guid_creation_num"].(string); ok {
		guid, err := strconv.ParseUint(guidStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid guid_creation_num %q: %w", guidStr, err)
		}
		core.GuidCreationNum = guid
	}

	return core, nil
}

// ObjectOwner returns the current owner of an object.
func ObjectOwner(ctx context.Context, client Client, object AccountAddress, opts ...ResourceOption) (AccountAddress, error) {
	core, err := GetObjectCore(ctx, client, object, opts...)
	if err != nil {
		return AccountAddress{}, err
	}
	return core.Owner, nil
}

// IsObjectOwner reports whether owner is the current owner of an object.
func IsObjectOwner(ctx context.Context, client Client, object AccountAddress, owner AccountAddress, opts ...ResourceOption) (bool, error) {
	actual, err := ObjectOwner(ctx, client, object, opts...)
	if err != nil {
		return false, err
	}
	return actual == owner, nil
}

// ObjectTransferPayload builds an entry-function payload that transfers an
// object to a new owner via 0x1::object::transfer, using 0x1::object::ObjectCore
// as the object type. This works for any transferable object; for typed objects
// that require a more specific type argument, use ObjectTransferPayloadOf.
func ObjectTransferPayload(object AccountAddress, to AccountAddress) *EntryFunctionPayload {
	return ObjectTransferPayloadOf(object, to, objectCoreTypeTag)
}

// ObjectTransferPayloadOf builds a 0x1::object::transfer payload with an
// explicit object type argument.
func ObjectTransferPayloadOf(object AccountAddress, to AccountAddress, objectType TypeTag) *EntryFunctionPayload {
	return &EntryFunctionPayload{
		Module:   ModuleID{Address: AccountOne, Name: "object"},
		Function: "transfer",
		TypeArgs: []TypeTag{objectType},
		Args:     []any{object, to},
	}
}
