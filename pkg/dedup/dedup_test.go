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
