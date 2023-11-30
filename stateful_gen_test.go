package ion

import "testing"

func TestStateGenSplitElem(t *testing.T) {
	i := 0
	seq := StateGen(func() (int, bool) {
		ret := i
		i += 1
		return ret, true
	})

	if e, ok := seq.Elem(0); ok == false || e != 0 {
		t.Fatalf("Expected seq[0] == 0, true, but was %d, %t", e, ok)
	}

	l, r := seq.Split(10)

	if e, ok := seq.Elem(0); ok == false || e != 0 {
		t.Fatalf("Expected seq[0] == 0, true, but was %d, %t", e, ok)
	}

	if e, ok := l.Elem(0); ok == false || e != 0 {
		t.Fatalf("Expected l[0] == 0, true, but was %d, %t", e, ok)
	}
	if e, ok := l.Elem(9); ok == false || e != 9 {
		t.Fatalf("Expected l[9] == 9, true, but was %d, %t", e, ok)
	}

	if e, ok := r.Elem(0); ok == false || e != 10 {
		t.Fatalf("Expected r[0] == 10, true, but was %d, %t", e, ok)
	}

	// Must be able to split a stateSplit
	ll, rr := r.Split(10)
	if e, ok := r.Elem(0); ok == false || e != 10 {
		// r should still give correct results even after a split
		t.Fatalf("Expected r[0] == 10, true, but was %d, %t", e, ok)
	}
	if e, ok := ll.Elem(0); ok == false || e != 10 {
		t.Fatalf("Expected ll[0] == 10, true, but was %d, %t", e, ok)
	}
	if e, ok := ll.Elem(9); ok == false || e != 19 {
		t.Fatalf("Expected ll[9] == 19, true, but was %d, %t", e, ok)
	}

	if e, ok := rr.Elem(0); ok == false || e != 20 {
		t.Fatalf("Expected rr[0] == 20, true, but was %d, %t", e, ok)
	}

}

func TestStateGenTakeElem(t *testing.T) {
	i := 0
	seq := StateGen(func() (int, bool) {
		ret := i
		i += 1
		return ret, true
	})

	if e, ok := seq.Elem(0); ok == false || e != 0 {
		t.Fatalf("Expected seq[0] == 0, true, but was %d, %t", e, ok)
	}

	l := seq.Take(10)

	if e, ok := seq.Elem(0); ok == false || e != 0 {
		t.Fatalf("Expected seq[0] == 0, true, but was %d, %t", e, ok)
	}

	if e, ok := l.Elem(0); ok == false || e != 0 {
		t.Fatalf("Expected l[0] == 0, true, but was %d, %t", e, ok)
	}
	if e, ok := l.Elem(9); ok == false || e != 9 {
		t.Fatalf("Expected l[9] == 9, true, but was %d, %t", e, ok)
	}

	l, r := seq.Split(10)

	// Must be able to Take from a stateSplit
	ll := r.Take(10)
	if e, ok := r.Elem(0); ok == false || e != 10 {
		// r should still give correct results even after a split
		t.Fatalf("Expected r[0] == 10, true, but was %d, %t", e, ok)
	}
	if e, ok := ll.Elem(0); ok == false || e != 10 {
		t.Fatalf("Expected ll[0] == 10, true, but was %d, %t", e, ok)
	}
	if e, ok := ll.Elem(9); ok == false || e != 19 {
		t.Fatalf("Expected ll[9] == 19, true, but was %d, %t", e, ok)
	}
}
