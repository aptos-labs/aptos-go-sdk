package aptos

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStructTag(t *testing.T) {
	st := StructTag{
		Address: AccountOne,
		Module:  "coin",
		Name:    "CoinStore",
		TypeParams: []TypeTag{
			TypeTag{Value: &StructTag{
				Address:    AccountOne,
				Module:     "aptos_coin",
				Name:       "AptosCoin",
				TypeParams: nil,
			}},
		},
	}
	assert.Equal(t, "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>", st.String())
	var aa3 AccountAddress
	aa3.ParseStringRelaxed("0x3")
	st.TypeParams = append(st.TypeParams, TypeTag{Value: &StructTag{
		Address:    aa3,
		Module:     "other",
		Name:       "thing",
		TypeParams: nil,
	}})
	assert.Equal(t, "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin,0x3::other::thing>", st.String())
}
