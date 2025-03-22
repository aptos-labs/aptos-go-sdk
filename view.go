package aptos

import (
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

// ViewPayload is a payload for a view function
type ViewPayload struct {
	Module   ModuleId  // ModuleId of the View function e.g. 0x1::coin
	Function string    // Name of the View function e.g. balance
	ArgTypes []TypeTag // TypeTags of the type arguments
	Args     [][]byte  // Arguments to the function encoded in BCS
}

func (vp *ViewPayload) MarshalBCS(ser *bcs.Serializer) {
	vp.Module.MarshalBCS(ser)
	ser.WriteString(vp.Function)
	bcs.SerializeSequence(vp.ArgTypes, ser)
	length, err := util.IntToU32(len(vp.Args))
	if err != nil {
		ser.SetError(err)
		return
	}
	ser.Uleb128(length)
	for _, a := range vp.Args {
		ser.WriteBytes(a)
	}
}
