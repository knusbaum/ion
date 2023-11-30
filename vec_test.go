package ion

import (
	"io"
	"math/rand"
	"os"
	"testing"
)

// func TestVecAppend(t *testing.T) {
// 	var s *Vec

// 	s = s.Append(1)
// 	s = s.Append(2)
// 	s = s.Append(3)
// 	spew.Dump(s)

// 	s = s.Append(4)
// 	s = s.Append(5)
// 	s = s.Append(6)
// 	spew.Dump(s)

// 	s = s.Append(7)
// 	s = s.Append(8)
// 	spew.Dump(s)

// 	s = s.Append(9)
// 	spew.Dump(s)
// 	for i := uint64(10); i < 17; i++ {
// 		s = s.Append(i)
// 	}
// 	spew.Dump(s)
// 	s = s.Append(17)
// 	fmt.Printf("BREAK\n")
// 	s.Dot(os.Stdout)

// 	for i := uint64(18); i <= 33; i++ {
// 		s = s.Append(i)
// 	}
// 	s.Dot(os.Stdout)
// }

// func TestVecBigJoin(t *testing.T) {
// 	var s *Vec[uint64]
// 	for i := uint64(0); i < 1800; i++ {
// 		//writeDot(s, "out.dot")
// 		var s2 *Vec[uint64]
// 		s2 = s2.Append(i)
// 		//writeDot(s2, "out2.dot")
// 		s = s2.Join(s)
// 		//writeDot(s, "out3.dot")
// 	}
// 	writeDot(s, "out3.dot")
// }

func TestVecAppend(t *testing.T) {
	var s *Vec[uint64]
	for i := uint64(0); i < 20; i++ {
		s = s.Append(i)
	}

	s2 := s
	for i := uint64(20); i < 40; i++ {
		s2 = s2.Append(i)
	}

	for i := uint64(0); i < 40; i++ {
		e, ok := s2.Elem(i)
		if !ok || e != i {
			t.Fatalf("Expected element at index %d to be %d, but got %d.\n", i, i, e)
		}
	}

	//s.Elem(30)
	if l := s.Len(); l != 20 {
		t.Fatalf("Expected s to have length 20, but has length %d\n", l)
	}

	if l := s2.Len(); l != 40 {
		t.Fatalf("Expected s2 to have length 40, but has length %d\n", l)
	}
}

// func TestVecSplit(t *testing.T) {
// 	var s *Vec[uint64]
// 	for i := uint64(0); i < 32; i++ {
// 		s = s.Append(i)
// 	}

// 	f, _ := os.Create("out2.dot")
// 	s.Dot(f)
// 	f.Close()

// 	s1, s2 := s.split(25)
// 	f, _ = os.Create("out3.dot")
// 	s1.Dot(f)
// 	f.Close()

// 	f, _ = os.Create("out4.dot")
// 	s2.Dot(f)
// 	f.Close()
// }

func TestVecJoin(t *testing.T) {
	var s1, s2 *Vec[uint64]

	for i := uint64(0); i < 20; i++ {
		s1 = s1.Append(i)
		s2 = s2.Append(i)
	}

	s3 := s1.Join(s2)

	if l := s3.Len(); l != 40 {
		t.Fatalf("Expected length of 40 but was %d\n", l)
	}

	// f, _ := os.Create("out2.dot")
	// s1.Dot(f)
	// f.Close()

	// f, _ = os.Create("out3.dot")
	// s2.Dot(f)
	// f.Close()

	// f, _ = os.Create("out4.dot")
	// s3.Dot(f)
	// f.Close()
}

func TestVecSplitJoinSingle(t *testing.T) {
	t.Run("prepend", func(t *testing.T) {
		var s *Vec[uint64]
		var sl []uint64
		for i := uint64(0); i < 20; i++ {
			s = s.Append(i)
			sl = append(sl, i)
		}

		v := func(s *Vec[uint64]) {
			if bad := validateVec(s); bad != nil {
				s.Dot(os.Stdout)
				t.Fatalf("Failed to validate sequence.\n")
			}
		}

		v(s)

		for i := 0; i < 400; i++ {
			l, r := s.split(1)
			s = r.Join(l)
			v(s)
			sl = append(sl[1:], sl[:1]...)
			sliceEqual(t, asSlice(s), sl)
		}
	})
	t.Run("append", func(t *testing.T) {
		var s *Vec[uint64]
		var sl []uint64
		for i := uint64(0); i < 20; i++ {
			s = s.Append(i)
			sl = append(sl, i)
		}

		v := func(s *Vec[uint64]) {
			if bad := validateVec(s); bad != nil {
				s.Dot(os.Stdout)
				t.Fatalf("Failed to validate sequence.\n")
			}
		}

		v(s)

		for i := 0; i < 400; i++ {
			l, r := s.split(19)
			s = r.Join(l)
			v(s)
			sl = append(sl[19:], sl[:19]...)
			sliceEqual(t, asSlice(s), sl)
		}
	})
}

func TestVecBuild(t *testing.T) {
	s := BuildVec[int](func(add func(e int)) {
		for i := 0; i < 100; i++ {
			add(i)
		}
	})
	//TODO: finish test
	s.Dot(io.Discard)
}

func TestVecSplitJoinRandom(t *testing.T) {
	rand.Seed(1000)
	var s *Vec[uint64]
	var sl []uint64
	for i := uint64(0); i < 20; i++ {
		s = s.Append(i)
		sl = append(sl, i)
	}

	v := func(s *Vec[uint64]) {
		if bad := validateVec(s); bad != nil {
			s.Dot(os.Stdout)
			t.Fatalf("Failed to validate sequence.\n")
		}
	}

	v(s)

	var sum uint64
	s.Iterate(func(i uint64) bool {
		sum += i
		return true
	})

	//fmt.Printf("SUM: %d\n", sum)
	//return

	for i := 0; i < 10000; i++ {
		//fmt.Printf("SEQ : %v\nSLICE: %v\n", asSlice(s), sl)
		//fmt.Printf("LOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOp\n")
		rnd := rand.Uint64() % 20
		//fmt.Printf("SPLITTING AT %d\n",
		//	rnd)
		//s.Dot(os.Stdout)
		//writeDot(s, "out.dot")
		l, r := s.split(rnd)
		sll := sl[:rnd]
		slr := sl[rnd:]
		sliceEqual(t, asSlice(l), sll)
		sliceEqual(t, asSlice(r), slr)

		//fmt.Printf("Validating L:\n")
		//l.Dot(os.Stdout)
		//writeDot(l, "out2.dot")
		v(l)
		//fmt.Printf("##########################\n")

		//fmt.Printf("Validating R:\n")
		//r.Dot(os.Stdout)
		//writeDot(r, "out3.dot")
		v(r)
		//fmt.Printf("##########################\n")

		//fmt.Printf("JOINING:\n")
		s = r.Join(l)
		sl = append(slr, sll...)
		sliceEqual(t, asSlice(s), sl)

		//fmt.Printf("Validating S:\n")
		//s.Dot(os.Stdout)
		//writeDot(s, "out4.dot")
		v(s)

		newsum := uint64(0)
		s.Iterate(func(j uint64) bool {
			newsum += j
			return true
		})
		if newsum != sum {
			t.Fatalf("On iteration %d: expected sum to be %d, but was %d\n", i, sum, newsum)
		}
	}
	//fmt.Printf("SEQ : %v\nSLICE: %v\n", asSlice(s), sl)
	//s.Dot(os.Stdout)
	//writeDot(s, "out.dot")
}

func sliceEqual(t *testing.T, s, s2 []uint64) {
	if len(s) != len(s2) {
		t.Fatalf("%v != %v", s, s2)
	}
	for i := range s {
		if s[i] != s2[i] {
			t.Fatalf("%v != %v at index %d", s, s2, i)
		}
	}
}

func asSlice(s *Vec[uint64]) []uint64 {
	var sl []uint64
	s.Iterate(func(i uint64) bool {
		sl = append(sl, i)
		return true
	})
	return sl
}

func writeDot(s *Vec[uint64], name string) {
	f, _ := os.Create(name)
	defer f.Close()
	s.Dot(f)
}

func TestVecLeafCut(t *testing.T) {
	s := newLeaf[uint64]()
	s.seq = append(s.seq, 1)
	s.seq = append(s.seq, 2)
	s.seq = append(s.seq, 3)
	s.seq = append(s.seq, 4)
	s.mutCutFront(2)

	if s.is[0] != 3 {
		t.Fatalf("cutting the leaf failed.")
	}
}
