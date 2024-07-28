package types

import (
	"github.com/aptos-labs/aptos-go-sdk/bcs"
)

// ModuleId the identifier for a module e.g. 0x1::coin
type ModuleId struct {
	Address AccountAddress
	Name    string
}

func (mod *ModuleId) MarshalBCS(ser *bcs.Serializer) {
	mod.Address.MarshalBCS(ser)
	ser.WriteString(mod.Name)
}

func (mod *ModuleId) UnmarshalBCS(des *bcs.Deserializer) {
	mod.Address.UnmarshalBCS(des)
	mod.Name = des.ReadString()
}
