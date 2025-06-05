package dedup

import "testing"

func TestDeduper(t *testing.T) {
	d := New(2)
	if d.Seen("a") {
		t.Fatalf("first time 'a' should be unseen")
	}
	if !d.Seen("a") {
		t.Fatalf("second time 'a' should be seen")
	}
	if d.Seen("b") {
		t.Fatalf("first time 'b' should be unseen")
	}
	if d.Seen("c") {
		t.Fatalf("first time 'c' should be unseen")
	}
	// 'a' should have been evicted (capacity 2)
	if d.Seen("a") {
		t.Fatalf("'a' should have been evicted and be unseen again")
	}
}

func TestDeduperZeroCapacity(t *testing.T) {
	d := New(0)
	if d.capacity != 1 {
		t.Fatalf("expected capacity 1 for zero input, got %d", d.capacity)
	}
	if d.Seen("test") {
		t.Fatalf("first time should be unseen")
	}
	if !d.Seen("test") {
		t.Fatalf("second time should be seen")
	}
}

func TestDeduperNegativeCapacity(t *testing.T) {
	d := New(-5)
	if d.capacity != 1 {
		t.Fatalf("expected capacity 1 for negative input, got %d", d.capacity)
	}
}
