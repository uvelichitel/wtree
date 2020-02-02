package wutarray

import (
	"github.com/uvelichitel/wtree"
	"github.com/uvelichitel/wtree/bitmap64"
)

type Maps []bitmap64.Bitmap64

type Layout = bitmap64.Bitmap64

func Huffman(w []int) (Layout, []int) {
	if len(w)&1 != 0 {
		w = append(w, 0)
	}
	q := NewQueue()
	ec := Elem{}
	lc := Layout{}
	dc := []int{}
	e := [2]Elem{}
	var w1, w2 int
	for {
		for i := 0; i < 2; i++ {
			if q.count == 0 {
				e[i] = Elem{Weight: w[len(w)-1] + w[len(w)-2], Layout: Layout{0}, D: []int{len(w) - 2, len(w) - 1}}
				w = w[:len(w)-2]
				continue
			} else {
				w1 = q.buf[q.head].Weight
			}
			if len(w) == 0 {
				e[i] = q.Pop()
				continue
			} else {
				w2 = w[0] + w[1]
			}
			if w1 > w2 {
				e[i] = Elem{Weight: w[len(w)-1] + w[len(w)-2], Layout: Layout{0}, D: []int{len(w) - 2, len(w) - 1}}
				w = w[:len(w)-2]
			} else {
				e[i] = q.Pop()
			}
		}
		lc, dc = Combine(e[1].Layout, e[0].Layout, e[1].D, e[0].D)
		ec = Elem{Weight: e[0].Weight + e[1].Weight, Layout: lc, D: dc}
		if q.count == 0 && len(w) == 0 {
			return ec.Layout, ec.D
		}
		q.Push(ec)
	}
}

// minQueueLen is smallest capacity that queue may have.
// Must be power of 2 for bitwise modulus: x % n == x & (n - 1).
const minQueueLen = 16

type Elem struct {
	Weight int
	Layout
	D []int
}

// Queue represents a single instance of the queue data structure.
type Queue struct {
	buf               []Elem
	head, tail, count int
}

// New constructs and returns a new Queue.
func NewQueue() *Queue {
	return &Queue{
		buf: make([]Elem, minQueueLen),
	}
}

// resizes the queue to fit exactly twice its current contents
// this can result in shrinking if the queue is less than half-full
func (q *Queue) resize() {
	newBuf := make([]Elem, q.count<<1)

	if q.tail > q.head {
		copy(newBuf, q.buf[q.head:q.tail])
	} else {
		n := copy(newBuf, q.buf[q.head:])
		copy(newBuf[n:], q.buf[:q.tail])
	}

	q.head = 0
	q.tail = q.count
	q.buf = newBuf
}

// Push() puts an element on the end of the queue.
func (q *Queue) Push(elem Elem) {
	if q.count == len(q.buf) {
		q.resize()
	}

	q.buf[q.tail] = elem
	// bitwise modulus
	q.tail = (q.tail + 1) & (len(q.buf) - 1)
	q.count++
}

// Pop removes and returns the element from the front of the queue. If the
// queue is empty, the call will panic.
func (q *Queue) Pop() Elem {
	if q.count <= 0 {
		panic("queue: Pop() called on empty queue")
	}
	ret := q.buf[q.head]
	q.buf[q.head] = Elem{}
	// bitwise modulus
	q.head = (q.head + 1) & (len(q.buf) - 1)
	q.count--
	// Resize down if buffer 1/4 full.
	if len(q.buf) > minQueueLen && (q.count<<2) == len(q.buf) {
		q.resize()
	}
	return ret
}

func Combine(l1, l2 Layout, d1, d2 []int) (Layout, []int) {
	d := make([]int, 0)
	l := make(Layout, 0)
	l.Push(1)
	var b int8
	for c1, c2, i := 1, 1, 0; c1 > 0 || c2 > 0; {
		for i, c1 = c1, 0; i > 0; i-- {
			if l1[0] == 1 { //TODO corner cases. Switch uint64 bitmap
				if len(l1) == 1 {
					break
				}
				l1 = l1[1:]
				b = int8(l1[0] & 1)
				l1[0] = l1[0] >> 1
				l1[0] = l1[0] + (1 << 63)
			} else {
				b = int8(l1[0] & 1)
				l1[0] = l1[0] >> 1
			}
			l.Push(b)
			if b == 0 {
				d = append(d, d1[:2]...)
				d1 = d1[2:]
			} else {
				c1 += 2
			}
		}
		for i, c2 = c2, 0; i > 0; i-- {
			if l2[0] == 1 {
				if len(l2) == 1 {
					break
				}
				l2 = l2[1:]
				b = int8(l2[0] & 1)
				l2[0] = l2[0] >> 1
				l2[0] = l2[0] + (1 << 63)
			} else {
				b = int8(l2[0] & 1)
				l2[0] = l2[0] >> 1
			}
			l.Push(b)
			if b == 0 {
				d = append(d, d2[:2]...)
				d2 = d2[2:]
			} else {
				c2 += 2
			}
		}
	}
	return l, d
}

type UWTArray struct {
	*Maps
	Mark int
	*Layout
}

func (wt UWTArray) BitMap() wtree.Bitmap {
	return (*wt.Maps)[wt.Mark]
}
func (wt UWTArray) LChild() wtree.WTree {
	if wt.Mark == 0 {
		wt.Mark = 1
	} else {
		wt.Mark = int((*wt.Layout).Rank1(uint(wt.Mark)-1)*2 + 1)
	}
	return wt
}
func (wt UWTArray) RChild() wtree.WTree {
	if wt.Mark == 0 {
		wt.Mark = 2
	} else {
		wt.Mark = int((*wt.Layout).Rank1(uint(wt.Mark)-1)*2 + 2)
	}
	return wt
}
func (wt UWTArray) Parrent() wtree.WTree {
	wt.Mark = int((*wt.Layout).Select1((uint(wt.Mark)+1)/2 - 1))
	return wt
}
func (wt UWTArray) IsLChild() bool {
	return wt.Mark&1 != 0
}
func (wt UWTArray) IsHead() bool {
	return wt.Mark == 0
}
func (wt UWTArray) IsLeaf() bool {
	return (*wt.Layout).Get(uint(wt.Mark)) == 0
}
func (wt UWTArray) FromLeaf(pos uint) int {
	bit := (*wt.Maps)[wt.Mark].Get(pos)
	return int(2*((*wt.Layout).Rank0(uint(wt.Mark-1)))) + int(bit)
}
func (wt *UWTArray) ToLeaf(ind int) int8 {
	bit := int8(ind & 1)
	wt.Mark = int((*wt.Layout).Select0(uint(ind) / 2))
	return bit
}
func (wta UWTArray) Access(pos uint) int {
	wt, pos := wtree.Access(wta, pos)
	wta = wt.(UWTArray)
	return wta.FromLeaf(pos)
}

func (wta UWTArray) Track(ind int, count uint) uint {
	bit := wta.ToLeaf(ind)
	_, pos := wtree.Track(wta, count, bit)
	return pos
}
func (wta UWTArray) Append(ind int) {
	b := wta.ToLeaf(ind)
	for {
		(*wta.Maps)[wta.Mark].Push(b)
		b = int8((wta.Mark - 1) & 1)
		wta = wta.Parrent().(UWTArray)
		if wta.IsHead() {
			(*wta.Maps)[wta.Mark].Push(b)
			return
		}
	}
}
