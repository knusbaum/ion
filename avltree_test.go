package ion

import (
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestAVLInsert(t *testing.T) {
	var tr *AVLTree[uint64, uint64]
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
			t.Fatalf("Expected to find %d in the tree, but did not.", i)
		}
		if j != i {
			t.Fatalf("Expected j == i, but j == %v", j)
		}
	}
}

// func TestAVLC1(t *testing.T) {

// 	var tr *AVLTree
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

func checkHeight[T Number, U any](t *testing.T, tr *AVLTree[T, U]) *AVLTree[T, U] {
	if tr == nil {
		return nil
	}
	if r := checkHeight(t, tr.l); r != nil {
		return r
	}
	if r := checkHeight(t, tr.r); r != nil {
		return r
	}

	var l, r int8

	if tr.l != nil {
		l = tr.l.height
	}
	if tr.r != nil {
		r = tr.r.height
	}

	if l > r {
		if tr.height != l+1 {
			//t.Fatalf("T: %v, L: %v, R: %v Expected node %d to have height %d, but had %d\n",
			//	tr, tr.l, tr.r, tr.i, l+1, tr.height)
			return tr
		}
	} else if tr.height != r+1 {
		//t.Fatalf("T: %v, L: %v, R: %v Expected node %d to have height %d, but had %d\n",
		//	tr, tr.l, tr.r, tr.i, r+1, tr.height)
		return tr
	}
	return nil
}

func checkBalance[T Number, U any](t *testing.T, tr *AVLTree[T, U]) *AVLTree[T, U] {
	if tr == nil {
		return nil
	}
	if r := checkBalance(t, tr.l); r != nil {
		return r
	}
	if r := checkBalance(t, tr.r); r != nil {
		return r
	}
	if b := tr.bf(); b >= 2 || b <= -2 {
		return tr
	}
	return nil
}

func FuzzAVLInsDel(f *testing.F) {
	f.Add(make([]byte, 9000))
	// The format of the byte array is a series of elements of 9 bytes:
	// [0][1][2][3][4][5][6][7][8]
	// Byte 0 % 2 determines if the operation is an insertion or a deletion
	// Bytes 1-8 are the 64-bit integer
	// Dangling bytes not totalling 9 bytes are excluded.
	f.Fuzz(func(t *testing.T, val []byte) {
		var tr *AVLTree[uint64, uint64]
		ks := make(map[uint64]struct{})
		dels := make(map[uint64]struct{})
		//inss := make(map[uint64]struct{})
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
			if _, ok := ks[v]; !ok {
				ntr := tr.Insert(v, v)
				ks[v] = struct{}{}
				sz++
				tr = ntr
				if del {
					// Store some of the values to delete later.
					//dels = append(dels, v)
					if _, ok := dels[v]; !ok {
						dels[v] = struct{}{}
						szd++
					}
				}
			}
		}

		otr := tr

		for k := range dels {
			s := tr.Size()
			ntr, ok := tr.Delete(k)
			if !ok {
				t.Fatalf("Expected to delete %d, but failed.", k)
			}
			tr = ntr
			if tr.Size() != s-1 {
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

func FuzzAVLConsistency(f *testing.F) {
	f.Add(make([]byte, 9000))
	// The format of the byte array is a series of elements of 9 bytes:
	// [0][1][2][3][4][5][6][7][8]
	// Byte 0 % 2 determines if the operation is an insertion or a deletion
	// Bytes 1-8 are the 64-bit integer
	// Dangling bytes not totalling 9 bytes are excluded.
	f.Fuzz(func(t *testing.T, val []byte) {
		var tr *AVLTree[uint64, uint64]
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
			if ftr := checkHeight(t, tr); ftr != nil {
				t.Fatalf("Bad height for node %#v\n", ftr)
			}
			if ftr := checkBalance(t, tr); ftr != nil {
				// f, _ := os.Create("after.dot")
				// tr.Dot(f)
				// f.Close()
				if insert {
					t.Logf("After insert of %d\n", v)
				} else {
					t.Logf("After delete of %d\n", v)
				}
				t.Fatalf("Bad balance for node %v\n", ftr)
			}
		}

	})
}

func TestAVLConsistency(t *testing.T) {
	t.Run("rand-insert-delete", func(t *testing.T) {
		seed := time.Now().UnixNano()
		//seed := int64(1700079657806063024)
		//fmt.Printf("SEED: %d\n", seed)
		rand := rand.New(rand.NewSource(seed))

		insert := 0
		del := 0
		var tr *AVLTree[uint64, uint64]
		var dels []uint64
		const sz = 100
		for i := 0; i < sz; i++ {
			r := rand.Uint64()
			if rand.Int()%2 == 0 {
				// Store some of the values to delete later.
				dels = append(dels, r)
				del++
			}
			tr = tr.Insert(r, r)
			insert++
		}

		if tr.Size() != sz {
			t.Fatalf("(Seed %d): Expected tree size to be %d, but was %d\n", seed, sz, tr.Size())
		}
		for _, k := range dels {
			s := tr.Size()
			tr, _ = tr.Delete(k)
			if tr.Size() != s-1 {
				t.Fatalf("(Seed %d): Wrong size after deleting %d. Expected %d, but got %d\n", seed, k, s-1, tr.Size())
			}
		}

		if tr.Size() != uint64(sz-len(dels)) {
			t.Fatalf("(Seed %d): Expected tree size to be %d, but was %d\n", seed, sz-len(dels), tr.Size())
		}
	})

	t.Run("height-balance", func(t *testing.T) {
		var tr *AVLTree[uint64, uint64]
		const sz = 100000
		checkEach := false

		//seed := int64(1700076327895896749)
		seed := time.Now().UnixNano()
		rand := rand.New(rand.NewSource(seed))

		//found := 0
		for i := 0; i < sz; i++ {
			// Use a smaller integer space to ensure we delete something eventually
			r := rand.Uint32()
			if rand.Intn(2) == 0 {
				tr, _ = tr.Delete(uint64(0))
				if checkEach {
					if ftr := checkHeight(t, tr); ftr != nil {
						ftr.Dot(os.Stdout)
						t.Fatalf("After delete (%d), T: %v, L: %v, R: %v Expected node %d to have height _, but had %d\n",
							r, ftr, ftr.l, ftr.r, ftr.k, ftr.height)
					}
					if ftr := checkBalance(t, tr); ftr != nil {
						ftr.Dot(os.Stdout)
						t.Fatalf("After delete (%d), T: %v, L: %v, R: %v Found a balance factor %d\n",
							r, ftr, ftr.l, ftr.r, ftr.bf())
					}
				}
			} else {
				tr = tr.Insert(uint64(r), uint64(r))
				if checkEach {
					if ftr := checkHeight(t, tr); ftr != nil {
						ftr.Dot(os.Stdout)
						t.Fatalf("After insert (%d), T: %v, L: %v, R: %v Expected node %d to have height _, but had %d\n",
							r, ftr, ftr.l, ftr.r, ftr.k, ftr.height)
					}
					if ftr := checkBalance(t, tr); ftr != nil {
						ftr.Dot(os.Stdout)
						t.Fatalf("After insert (%d), T: %v, L: %v, R: %v Found a balance factor %d\n",
							r, ftr, ftr.l, ftr.r, ftr.bf())
					}
				}
			}

			if i%1000 == 0 {
				//fmt.Printf("I: %d, Size: %d, Height: %d\n", i, tr.Size(), height(tr))
				if ftr := checkHeight(t, tr); ftr != nil {
					ftr.Dot(os.Stdout)
					t.Fatalf("Seed (%d): T: %v, L: %v, R: %v Expected node %d to have height _, but had %d\n",
						seed, ftr, ftr.l, ftr.r, ftr.k, ftr.height)
				}
				if ftr := checkBalance(t, tr); ftr != nil {
					ftr.Dot(os.Stdout)
					t.Fatalf("Seed (%d): T: %v, L: %v, R: %v Found a balance factor %d\n",
						seed, ftr, ftr.l, ftr.r, ftr.bf())
				}
			}

		}
	})

}

func TestAVLDelete(t *testing.T) {
	var tr *AVLTree[uint64, uint64]
	for i := uint64(1); i <= 50; i++ {
		tr = tr.Insert(i, i)
	}

	tr, _ = tr.Delete(10)

	for i := uint64(1); i <= 9; i++ {
		j, ok := tr.Get(i)
		if !ok {
			t.Fatalf("Expected to find %d in the tree, but did not.", i)
		}
		if j != i {
			t.Fatalf("Expected j == i, but j == %v", j)
		}
	}
	j, ok := tr.Get(10)
	if ok {
		t.Errorf("Expected NOT to find 10 in the tree, but found it.")
	}
	if j != 0 {
		t.Error("Expected j == 0")
	}

}

// func TestAVLRotate(t *testing.T) {
// 	a := &AVLTree{i: 1}
// 	b := &AVLTree{i: 2}
// 	c := &AVLTree{i: 3}
// 	x := &AVLTree{i: 9, l: a, r: b}
// 	y := &AVLTree{i: 10, l: x, r: c}

// 	newY := y.right_rotate()
// 	if newY.i != x.i {
// 		t.Errorf("Expected x to be new root.\n")
// 	}
// 	if x.l.i != a.i {
// 		t.Errorf("Expected x.l == a\n")
// 	}
// 	if x.r.i != y.i {
// 		t.Errorf("Expected x.r == y\n")
// 	}
// 	if y.l.i != b.i {
// 		t.Errorf("Expected y.l == b\n")
// 	}
// 	if y.r.i != c.i {
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

func BenchmarkAVLTree(b *testing.B) {
	var t *AVLTree[uint64, uint64]

	for i := 0; i < b.N; i++ {
		r := rand.Uint64()
		t = t.Insert(r, r)
	}
}
