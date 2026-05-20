package confidentialasset

import (
	"testing"
)

func Test_viewToJSONArray_topLevelArray(t *testing.T) {
	t.Parallel()
	arr, err := viewToJSONArray([]any{true})
	if err != nil {
		t.Fatal(err)
	}
	if len(arr) != 1 {
		t.Fatalf("len=%d", len(arr))
	}
	b, ok := arr[0].(bool)
	if !ok || !b {
		t.Fatalf("got %#v", arr[0])
	}
}

func Test_viewToJSONArray_multipleBools(t *testing.T) {
	t.Parallel()
	arr, err := viewToJSONArray([]any{false, true})
	if err != nil {
		t.Fatal(err)
	}
	if len(arr) != 2 {
		t.Fatalf("len=%d", len(arr))
	}
}

func Test_viewBool(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		in   any
		want bool
	}{
		{true, true},
		{false, false},
		{"true", true},
		{"false", false},
	} {
		got, err := viewBool(tc.in)
		if err != nil {
			t.Fatalf("%#v: %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("%#v: got %v want %v", tc.in, got, tc.want)
		}
	}
	if _, err := viewBool(3); err == nil {
		t.Fatal("expected error for int")
	}
}

func Test_viewToJSONArray_singleObject(t *testing.T) {
	t.Parallel()
	arr, err := viewToJSONArray([]any{map[string]any{"k": true}})
	if err != nil {
		t.Fatal(err)
	}
	if len(arr) != 1 {
		t.Fatalf("len=%d", len(arr))
	}
}

func Test_viewToJSONArray_invalid(t *testing.T) {
	t.Parallel()
	ch := make(chan int)
	_, err := viewToJSONArray([]any{ch})
	if err == nil {
		t.Fatal("expected marshal error")
	}
}
