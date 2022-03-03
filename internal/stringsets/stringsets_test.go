package stringsets

import (
	"testing"
)

func TestCycle(t *testing.T) {
	s := New()
	if s.Len() != 0 {
		t.Fatalf("New set not empty")
	}

	s.Add("a")
	if !s.Has("a") {
		t.Fatalf("Basic creation and addition failed")
	}

	if s.Len() != 1 {
		t.Fatalf("Set wrong length after addition")
	}

	s.Discard("a")
	if s.Has("a") {
		t.Fatalf("Set still has element after discard")
	}
	if s.Len() != 0 {
		t.Fatalf("Set wrong length after discard")
	}
}

func TestMulti(t *testing.T) {
	s := New()
	s.Add("a")
	s.Add("b")
	if s.Len() != 2 {
		t.Fatalf("Set wrong length after two adds")
	}
	s.Discard("b")
	if s.Len() != 1 {
		t.Fatalf("Set wrong length after two adds, 1 discard")
	}
	s.Add("b")
	if s.Len() != 2 {
		t.Fatalf("Set wrong length after 2a1d1a")
	}
	s2 := s.Copy()
	if s2.Len() != 2 {
		t.Fatalf("Set copy wrong length")
	}
	s.Clear()
	if s.Len() != 0 {
		t.Fatalf("Set wrong length after clear")
	}
}

func TestSliceOps(t *testing.T) {
	s := New()
	s.AddSlice([]string{"a", "b", "c", "d"})
	if s.Len() != 4 {
		t.Fatalf("Set wrong length after adding slice")
	}
	s.DiscardSlice([]string{"b", "c"})
	if s.Len() != 2 {
		t.Fatalf("Set wrong length after discarding slice")
	}
	if !s.EqualsSlice([]string{"a", "d"}) {
		t.Fatalf("Set has wrong contents after discarding slice")
	}
}

func TestDiff(t *testing.T) {
	s1 := New()
	s2 := New()
	s1.Add("a")
	s1.Add("b")
	s1.Add("c")
	s2.Add("b")
	s3 := s1.Difference(s2)
	if s3.Len() != 2 {
		t.Fatalf("Set 3 wrong length after difference")
	}
	if !(s3.Has("a") && !s3.Has("b") && s3.Has("c")) {
		t.Fatalf("Set 3 has wrong contents after difference")
	}
}
