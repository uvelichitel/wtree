package dict

import (
	"encoding/binary"
	"errors"
	"io"
	"sort"
	"strings"
)

type mark int8

const (
	match mark = 1 << iota
	leaf
	tail
)

type Tree []node

type node struct {
	value int
	next  int
	label byte
	check mark
}

type kv struct {
	k []byte
	v int
}

// New builds new Tree from terms.
func newt(terms []string) Tree {
	kvs := make([]kv, 0, len(terms))
	for i, k := range terms {
		kvs = append(kvs, kv{[]byte(k), i})
	}
	sort.Slice(kvs, func(i, j int) bool {
		a, b := kvs[i].k, kvs[j].k
		for i := 0; i < len(a) && i < len(b); i++ {
			if a[i] == b[i] {
				continue
			}
			return a[i] < b[i]
		}
		return len(a) < len(b)
	})

	t := Tree{node{next: 1}}

	t = t.construct(kvs, 0, 0)
	return t
}

func (t Tree) construct(kvs []kv, depth, current int) Tree {
	if depth == len(kvs[0].k) || t[current].check&tail != 0 {
		t[current].check += match
		t[current].value = kvs[0].v
		kvs = kvs[1:]
		if len(kvs) == 0 || t[current].check&tail != 0 {
			t[current].check += leaf
			return t
		}
	}
	p := []int{0}
	for i := 0; i < len(kvs); {
		t = append(t, node{
			label: kvs[i].k[depth],
		})
		for c := kvs[i].k[depth]; i < len(kvs) && kvs[i].k[depth] == c; i++ {
		}

		if i == 1 && len(kvs[0].k) >= depth+1 { //TODO
			t[t.nextOf(current)].check += tail
		}

		p = append(p, i)
	}
	for i := 0; i < len(p)-1; i++ {
		t[t.nextOf(current)+i].next = len(t) - t.nextOf(current) - i
		t = t.construct(kvs[p[i]:p[i+1]], depth+1, t.nextOf(current)+i)
	}
	return t
}

// Trace returns the subtree of t whose root is the node traced from the root
// of t by path. It doesn't modify t itself, but returns the subtree.
func (t Tree) Trace(path []byte, d []string) Tree {
	if t == nil {
		return nil
	}

	var u int
	for _, c := range path {
		if t[u].check&leaf != 0 {
			return nil
		}
		u = t.nextOf(u)
		v := t.nextOf(u)
		if v-u > 40 {
			// Binary Search
			u += sort.Search(v-u, func(m int) bool {
				return t[u+m].label >= c
			})
		} else {
			// Linear Search
			for ; u != v-1 && t[u].label < c; u++ {
			}
		}

		if t[u].check&tail != 0 { //TODO
			if strings.HasPrefix(d[t[u].value], string(path)) {
				return t[u:]
			}
		}

		if u > v || t[u].label != c {
			return nil
		}

	}
	return t[u:]
}

func (t Tree) TraceByte(c byte, d []string) Tree {
	return t.Trace([]byte{c}, d)
}

// Terminal returns the value of the root of t. The second return value
// indicates whether the node has a value; if it is false, the first return
// value is nil. It returns nil also when the t is nil.
func (t Tree) Terminal() (int, bool) {
	if len(t) == 0 {
		return 0, false
	}
	return t[0].value, t[0].check&match != 0
}

// Predict returns the all values in the tree, t. The complexity is proportional
// to the number of nodes in t(it's not equal to len(t)).
func (t Tree) Predict() []int {
	if len(t) == 0 || t[0].check&leaf != 0 {
		return nil
	}

	// Search linearly all of the child.
	var end int
	for t[end].check&leaf == 0 {
		end = t.nextOf(t.nextOf(end)) - 1
	}

	var values []int
	for i := t.nextOf(0); i <= end; i++ {
		if t[i].check&match != 0 {
			values = append(values, t[i].value)
		}
	}
	return values
}

// Children returns the bytes of all direct children of the root of t. The result
// is sorted in ascending order.
func (t Tree) Children() []byte {
	if len(t) == 0 || t[0].check&leaf != 0 {
		return nil
	}

	var children []byte
	for _, c := range t[t.nextOf(0):t.nextOf(t.nextOf(0))] {
		children = append(children, c.label)
	}
	return children
}

// nextOf returns the index of the next node of t[i].
func (t Tree) nextOf(i int) int {
	return i + t[i].next
}

type Dict struct {
	terms []string
	Tree
}

func (d Dict) Serialize(w io.Writer) (err error) {
	j := strings.Join(d.terms, " ")
	l := int64(len(j))
	err = binary.Write(w, binary.LittleEndian, l)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, []byte(j))
	if err != nil {
		return err
	}
	l = int64(len(d.Tree))
	err = binary.Write(w, binary.LittleEndian, l)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, d.Tree)
	return err
}

func Restore(r io.Reader) (d Dict, err error) {
	l := new(int64)
	err = binary.Read(r, binary.LittleEndian, l)
	if err != nil {
		return d, err
	}
	j := make([]byte, *l)
	err = binary.Read(r, binary.LittleEndian, j)
	if err != nil {
		return d, err
	}
	d.terms = strings.Fields(string(j))
	err = binary.Read(r, binary.LittleEndian, l)
	if err != nil {
		return d, err
	}
	tree := make(Tree, *l)
	err = binary.Read(r, binary.LittleEndian, tree)
	if err != nil {
		return d, err
	}
	d.Tree = tree
	return d, err
}

func New(s []string) (d Dict) {
	d.terms = s
	d.Tree = newt(s)
	return
}

func (d Dict) Get(i int) string { //TODO errors
	return d.terms[i]
}

func (d Dict) Lookup(s string) (int, error) {
	v, ok := d.Tree.Trace([]byte(s), d.terms).Terminal()
	if !ok {
		return 0, errors.New("Term absent in dictionary")
	} else {
		return v, nil
	}
}
