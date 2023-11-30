package ion

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
)

func FuzzRBConsistency(f *testing.F) {
	f.Add(make([]byte, 9000))
	// The format of the byte array is a series of elements of 9 bytes:
	// [0][1][2][3][4][5][6][7][8]
	// Byte 0 % 2 determines if the operation is an insertion or a deletion
	// Bytes 1-8 are the 64-bit integer
	// Dangling bytes not totalling 9 bytes are excluded.
	f.Fuzz(func(t *testing.T, val []byte) {
		var tr *RBTree[uint64, uint64]
		for len(val) >= 9 {
			insert := val[0]%2 == 0
			v := uint64(val[1]) |
				(uint64(val[2]) << 8) |
				(uint64(val[3]) << 16) |
				(uint64(val[4]) << 24) |
				(uint64(val[5]) << 32) |
				(uint64(val[6]) << 40) |
				(uint64(val[7]) << 48) |
				(uint64(val[8]) << 56)
			val = val[9:]

			if insert {
				tr = tr.Insert(v, v)
			} else {
				tr, _ = tr.Delete(v)
			}
			if n := validateRBTree(tr); n != nil {
				tr.Dot(os.Stdout)
				t.Fatalf("Failed to validate tree: %v\n", n)
			}
			// if ftr := checkHeight(t, tr); ftr != nil {
			// 	t.Fatalf("Bad height for node %#v\n", ftr)
			// }
			// if ftr := checkBalance(t, tr); ftr != nil {
			// 	// f, _ := os.Create("after.dot")
			// 	// tr.Dot(f)
			// 	// f.Close()
			// 	if insert {
			// 		t.Logf("After insert of %d\n", v)
			// 	} else {
			// 		t.Logf("After delete of %d\n", v)
			// 	}
			// 	t.Fatalf("Bad balance for node %v\n", ftr)
			// }
		}

	})
}

func FuzzRBInsDel(f *testing.F) {
	f.Add(make([]byte, 9000))
	// The format of the byte array is a series of elements of 9 bytes:
	// [0][1][2][3][4][5][6][7][8]
	// Byte 0 % 2 determines if the operation is an insertion or a deletion
	// Bytes 1-8 are the 64-bit integer
	// Dangling bytes not totalling 9 bytes are excluded.
	f.Fuzz(func(t *testing.T, val []byte) {
		var tr *RBTree[uint64, uint64]
		ks := make(map[uint64]struct{})
		var dels []uint64
		sz := 0
		szd := 0
		for len(val) >= 9 {
			del := val[0]%2 == 0
			v := uint64(val[1]) |
				(uint64(val[2]) << 8) |
				(uint64(val[3]) << 16) |
				(uint64(val[4]) << 24) |
				(uint64(val[5]) << 32) |
				(uint64(val[6]) << 40) |
				(uint64(val[7]) << 48) |
				(uint64(val[8]) << 56)
			val = val[9:]
			s := tr.Size()
			if _, ok := ks[v]; !ok {
				ntr := tr.Insert(v, v)
				ks[v] = struct{}{}
				sz++
				tr = ntr
				if tr.Size() != s+1 {
					tr.Dot(os.Stdout)
					t.Fatalf("Wrong size after inserting %d. Expected %d, but got %d\n",
						v, s+1, tr.Size())
				}
				if del {
					// Store some of the values to delete later.
					dels = append(dels, v)
					szd++
				}
			}
		}
		//fmt.Printf("Pre-deletes\n")
		//tr.Dot(os.Stdout)

		otr := tr
		for _, k := range dels {
			s := tr.Size()
			ntr, ok := tr.Delete(k)
			if !ok {
				fmt.Printf("Before delete %d\n", k)
				tr.Dot(os.Stdout)
				fmt.Printf("After delete %d\n", k)
				ntr.Dot(os.Stdout)
				t.Fatalf("Expected to delete %d, but failed.", k)
			}
			tr = ntr
			//fmt.Printf("After delete %d\n", k)
			//tr.Dot(os.Stdout)
			if tr.Size() != s-1 {
				tr.Dot(os.Stdout)
				t.Fatalf("Wrong size after deleting %d. Expected %d, but got %d\n",
					k, s-1, tr.Size())
			}
		}

		if tr.Size() != uint64(sz-len(dels)) {
			t.Fatalf("Expected tree size to be %d, but was %d\n",
				sz-len(dels), tr.Size())
		}
		if otr.Size() != uint64(sz) {
			t.Fatalf("Original tree lost some elements during delete.\n")
		}
	})
}

// func TestRBC1(t *testing.T) {

// 	var tr *RBTree
// 	tr1, _ := tr.Insert(1)
// 	tr2, _ := tr1.Insert(2)
// 	tr3, _ := tr2.Insert(3)
// 	tr4, _ := tr3.Insert(4)
// 	tr5, _ := tr4.Insert(5)

// 	tr6, _ := tr5.Delete(1)
// 	tr7, _ := tr6.Delete(2)
// 	tr8, _ := tr7.Delete(3)
// 	tr9, _ := tr8.Delete(4)
// 	tr10, _ := tr9.Delete(5)

// 	tr1.Dot(os.Stdout)
// 	tr2.Dot(os.Stdout)
// 	tr3.Dot(os.Stdout)
// 	tr4.Dot(os.Stdout)
// 	tr5.Dot(os.Stdout)

// 	tr6.Dot(os.Stdout)
// 	tr7.Dot(os.Stdout)
// 	tr8.Dot(os.Stdout)
// 	tr9.Dot(os.Stdout)
// 	tr10.Dot(os.Stdout)
// }

// func TestRBC2(t *testing.T) {
// 	var tr *RBTree
// 	tr, _ = tr.Insert(3544385890265608240)
// 	tr, _ = tr.Insert(3616443484303536176)
// 	tr, _ = tr.Insert(3472328296227680304)
// 	fmt.Printf("PRE:\n")
// 	tr.Dot(os.Stdout)
// 	tr, _ = tr.Delete(3544385890265608240)
// 	//REMOVE LARGEST!
// 	//DELETE 3616443484303536176
// 	//DELETE 3472328296227680304

// 	tr.Dot(os.Stdout)

// 	if n := validateRBTree(tr); n != nil {
// 		tr.Dot(os.Stdout)
// 		t.Fatalf("Failed to validate tree: %v\n", n)
// 	}
// }

func TestRBInsert(t *testing.T) {
	var tr *RBTree[uint64, uint64]
	for i := uint64(0); i <= 5000; i++ {
		ntr := tr.Insert(i, i)
		if tr.Size() != i {
			t.Fatalf("Inserting %d: tree was modified.\n", i)
		}
		tr = ntr
	}

	for i := uint64(0); i <= 5000; i++ {
		j, ok := tr.Get(i)
		if !ok {
			t.Errorf("Expected to find %d in the tree, but did not.", i)
		}
		if j != i {
			t.Error("Expected j == i")
		}
	}
}

func TestRBTree(t *testing.T) {
	var tr *RBTree[uint64, uint64]
	for i := uint64(1); i <= 50; i++ {
		tr = tr.Insert(i, i)
	}

	if n := validateRBTree(tr); n != nil {
		tr.Dot(os.Stdout)
		t.Fatalf("Failed to validate tree: %v\n", n)
	}

	for i := uint64(1); i <= 50; i++ {
		j, ok := tr.Get(i)
		if !ok {
			t.Errorf("Expected to find %d in the tree, but did not.", i)
		}
		if j != i {
			t.Error("Expected j == i")
		}
	}
}

// func TestRBRotate(t *testing.T) {
// 	a := &RBTree{i: 1}
// 	b := &RBTree{i: 2}
// 	c := &RBTree{i: 3}
// 	x := &RBTree{i: 9, l: a, r: b}
// 	y := &RBTree{i: 10, l: x, r: c}

// 	newY := y.right_rotate()
// 	if newY != x {
// 		t.Errorf("Expected x to be new root.\n")
// 	}
// 	if x.l != a {
// 		t.Errorf("Expected x.l == a\n")
// 	}
// 	if x.r != y {
// 		t.Errorf("Expected x.r == y\n")
// 	}
// 	if y.l != b {
// 		t.Errorf("Expected y.l == b\n")
// 	}
// 	if y.r != c {
// 		t.Errorf("Expected y.r == c\n")
// 	}

// 	newX := newY.left_rotate()
// 	if newX != y {
// 		t.Errorf("Expected x to be new root.\n")
// 	}
// 	if x.l != a {
// 		t.Errorf("Expected x.l == a\n")
// 	}
// 	if x.r != b {
// 		t.Errorf("Expected x.r == b\n")
// 	}
// 	if y.l != x {
// 		t.Errorf("Expected y.l == x\n")
// 	}
// 	if y.r != c {
// 		t.Errorf("Expected y.r == c\n")
// 	}
// }

func BenchmarkRBTree(b *testing.B) {
	var t *RBTree[uint64, uint64]

	for i := 0; i < b.N; i++ {
		r := rand.Uint64()
		t = t.Insert(r, r)
	}
}

func TestRBTreeDelete(t *testing.T) {
	var tr *RBTree[uint64, uint64]

	for i := uint64(1); i < 20; i++ {
		tr = tr.Insert(i, i)
	}
	if n := validateRBTree(tr); n != nil {
		tr.Dot(os.Stdout)
		t.Fatalf("Failed to validate tree: %v\n", n)
	}
	//tr.Dot(os.Stdout)

	tr, _ = tr.Delete(1)

	if n := validateRBTree(tr); n != nil {
		tr.Dot(os.Stdout)
		t.Fatalf("Failed to validate tree: %v\n", n)
	}

	//tr.Dot(os.Stdout)

}

func blackDepth(t *RBTree[uint64, uint64], count int) int {
	if t.c == black {
		count++
	}
	if t.l != nil {
		return blackDepth(t.l, count)
	}
	if t.r != nil {
		return blackDepth(t.r, count)
	}
	return count
}

func validateRBTree(t *RBTree[uint64, uint64]) *RBTree[uint64, uint64] {
	if t == nil {
		return nil
	}
	depth := blackDepth(t, 0)
	//fmt.Printf("EXPECTING DEPTH %d\n", depth)
	return validateRBTreer(t, 0, depth)
}

func validateRBTreer(t *RBTree[uint64, uint64], count, depth int) *RBTree[uint64, uint64] {
	if t.c == black {
		count++
		//fmt.Printf("%d count -> %d\n", t.i, count)
	} else {
		if t.l != nil && t.r == nil ||
			t.l == nil && t.r != nil {
			//fmt.Printf("Found red node %v with 1 child\n", t)
			return t
		}
	}
	if t.l != nil {
		if n := validateRBTreer(t.l, count, depth); n != nil {
			return n
		}
	}
	if t.r != nil {
		if n := validateRBTreer(t.r, count, depth); n != nil {
			return n
		}
	}
	if t.r == nil && t.l == nil {
		if count != depth {
			//fmt.Printf("Found black node %v at wrong depth. Depth was %d, expected %d\n", t, count, depth)
			return t
		}
	}
	return nil
}
