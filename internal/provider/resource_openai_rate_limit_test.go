package provider

import "testing"

func TestInt64PointerValue(t *testing.T) {
	if got := int64PointerValue(nil); !got.IsNull() {
		t.Fatalf("int64PointerValue(nil).IsNull() = false, want true")
	}

	zero := 0
	if got := int64PointerValue(&zero); got.IsNull() || got.ValueInt64() != 0 {
		t.Fatalf("int64PointerValue(0) = %v, want 0", got)
	}

	value := 25
	if got := int64PointerValue(&value); got.IsNull() || got.ValueInt64() != 25 {
		t.Fatalf("int64PointerValue(25) = %v, want 25", got)
	}
}
