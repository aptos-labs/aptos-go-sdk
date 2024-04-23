package aptos

import "testing"

func TestAccountSpecialString(t *testing.T) {
	var aa AccountAddress
	aa[31] = 3
	aas := aa.String()
	if aas != "0x3" {
		t.Errorf("wanted 0x3 got %s", aas)
	}

	var aa2 AccountAddress
	err := aa2.ParseStringRelaxed("0x3")
	if err != nil {
		t.Errorf("unexpected err %s", err)
	}
	if aa2 != aa {
		t.Errorf("aa2 != aa")
	}
}
