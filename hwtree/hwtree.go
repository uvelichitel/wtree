package hwtree

import (
	"encoding/binary"
	"io"

	"github.com/uvelichitel/wtree/bitmap64"
	//"github.com/uvelichitel/wtree/hwtree"
)

type Maps []bitmap64.Bitmap64

type Layout = bitmap64.Bitmap64

//Huffman coded tree implemented with queue.
func Huffman(w []uint) (Layout, []int) {
	if len(w)&1 != 0 {
		w = w[:len(w)-1]
	}
	q := NewQueue()
	ec := Elem{}
	lc := Layout{}
	dc := []int{}
	e := [2]Elem{}
	var w1, w2 uint
	for {
		for i := 0; i < 2; i++ {
			if q.count == 0 {
				e[i] = Elem{Weight: w[len(w)-1] + w[len(w)-2], Layout: Layout{2}, D: []int{len(w) - 2, len(w) - 1}}
				w = w[:len(w)-2]
				continue
			} else {
				w1 = q.buf[q.head].Weight
			}
			if len(w) == 0 {
				e[i] = q.Pop()
				continue
			} else {
				w2 = w[len(w)-1] + w[len(w)-2]
			}
			if w2 < w1 {
				e[i] = Elem{Weight: w[len(w)-1] + w[len(w)-2], Layout: Layout{2}, D: []int{len(w) - 2, len(w) - 1}}
				w = w[:len(w)-2]
			} else {
				e[i] = q.Pop()
			}
		}
		lc, dc = Combine(e[0].Layout, e[1].Layout, e[0].D, e[1].D)
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
	Weight uint
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

//Huffman coded tree implemented with heap.
func Huffman1(w []uint) (Layout, []int) {
	if len(w)&1 != 0 {
		w = w[:len(w)-1]
	}
	var l Layout
	var d []int
	var e, e1, e2 Elem
	heap := make(Heap, 0)
	for i := 0; i < len(w)-1; i += 2 {
		heap.Push(Elem{Weight: w[i] + w[i+1], Layout: Layout{2}, D: []int{i, i + 1}})
	}
	for {
		e1 = heap.Pop()
		e2 = heap.Pop()
		l, d = Combine(e1.Layout, e2.Layout, e1.D, e2.D)
		if len(heap) == 0 {
			return l, d
		}
		e = Elem{Weight: e1.Weight + e2.Weight, Layout: l, D: d}
		heap.Push(e)
	}
}

type Heap []Elem

func (h *Heap) Pop() Elem {
	n := len(*h) - 1
	(*h)[0], (*h)[n] = (*h)[n], (*h)[0]
	down(*h, 0, n)
	item := (*h)[n]
	(*h)[n] = Elem{} // avoid memory leak
	*h = (*h)[0:n]
	return item
}
func (h *Heap) Push(e Elem) {
	(*h) = append(*h, e)
	up(*h, len(*h)-1)
}
func up(h Heap, j int) {
	for {
		i := (j - 1) / 2 // parent
		//if i == j || !(h[j].Weight < h[i].Weight) {
		if i == j || h[j].Weight > h[i].Weight {
			break
		}
		h[i], h[j] = h[j], h[i]
		j = i
	}
}

func down(h Heap, i0, n int) {
	for {
		j1 := 2*i0 + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && h[j2].Weight < h[j1].Weight {
			j = j2 // = 2*i + 2  // right child
		}
		if !(h[j].Weight < h[i0].Weight) {
			break
		}
		h[i0], h[j] = h[j], h[i0]
		i0 = j
	}
}

func Combine(l1, l2 Layout, d1, d2 []int) (Layout, []int) {
	d := make([]int, 0)
	l := make(Layout, 0)
	l.Push(1)
	var b int8
	for c1, c2, i := 1, 1, 0; c1 > 0 || c2 > 0; {
		for i, c1 = c1, 0; i > 0; i-- {
			if l1[0] == 1 {
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

type HWTArray struct {
	Maps
	Mark int
	Layout
}

func New(l Layout) (h HWTArray) {
	h.Maps = make(Maps, l.Length()-1)
	h.Layout = l
	return
}

//Format Layout length/Layout/Maps length/map[1] length/ map[1]...
func (wt HWTArray) Serialize(w io.Writer) (err error) {
	l := int64(len(wt.Layout))
	err = binary.Write(w, binary.LittleEndian, l)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, wt.Layout)
	if err != nil {
		return err
	}
	l = int64(len(wt.Maps))
	err = binary.Write(w, binary.LittleEndian, l)
	if err != nil {
		return err
	}
	for _, v := range wt.Maps {
		l = int64(len(v))
		err = binary.Write(w, binary.LittleEndian, l)
		if err != nil {
			return err
		}
		err = binary.Write(w, binary.LittleEndian, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func Restore(r io.Reader) (wt HWTArray, err error) {
	l:=new(int64)
	err = binary.Read(r, binary.LittleEndian, l)
	if err != nil {
		return wt, err
	}
	layout := make(Layout, *l)
	err = binary.Read(r, binary.LittleEndian, layout)
	if err != nil {
		return wt, err
	}
	err = binary.Read(r, binary.LittleEndian, l)
	if err != nil {
		return wt, err
	}
	maps := make(Maps, *l)
	for k, _ := range maps {
		err = binary.Read(r, binary.LittleEndian, l)
		if err != nil {
			return wt, err
		}
		maps[k] = make(bitmap64.Bitmap64, *l)
		err = binary.Read(r, binary.LittleEndian, maps[k])
		if err != nil {
			return wt, err
		}

	}
	wt.Layout=layout
	wt.Maps=maps
	return wt, err
}

func (wt HWTArray) Access(pos uint) int {
	l := wt.Layout
	bit := wt.Maps[0].Get(pos)
	if bit == 0 {
		pos = wt.Maps[0].Rank0(pos) - 1
		wt.Mark = 1
	} else {
		pos = wt.Maps[0].Rank1(pos) - 1
		wt.Mark = 2
	}
	for bit = wt.Maps[wt.Mark].Get(pos); l.Get(uint(wt.Mark)) != 0; bit = wt.Maps[wt.Mark].Get(pos) {
		if bit == 0 {
			pos = wt.Maps[wt.Mark].Rank0(pos) - 1
			wt.Mark = int(wt.Layout.Rank1(uint(wt.Mark)-1)*2 + 1)
		} else {
			pos = wt.Maps[wt.Mark].Rank1(pos) - 1
			wt.Mark = int(wt.Layout.Rank1(uint(wt.Mark)-1)*2 + 2)
		}
	}
	return int(2*(wt.Layout.Rank0(uint(wt.Mark-1)))) + int(bit)
}

func (wt HWTArray) Track(ind int, count uint) uint {
	wt.Mark = int(wt.Layout.Select0(uint(ind) / 2))
	lch := (ind + 1) & 1
	for {
		if lch != 0 {
			count = wt.Maps[wt.Mark].Select0(count)
		} else {
			count = wt.Maps[wt.Mark].Select1(count)
		}
		if wt.Mark == 0 {
			break
		}
		lch = wt.Mark & 1
		wt.Mark = int(wt.Layout.Select1((uint(wt.Mark)+1)/2 - 1))
	}
	return count
}

func (wt HWTArray) Append(ind int) {
	b := int8(ind & 1)
	wt.Mark = int(wt.Layout.Select0(uint(ind) / 2))
	for {
		wt.Maps[wt.Mark].Push(b)
		b = int8((wt.Mark - 1) & 1)
		wt.Mark = int(wt.Layout.Select1((uint(wt.Mark)+1)/2 - 1))
		if wt.Mark == 0 {
			wt.Maps[wt.Mark].Push(b)
			return
		}
	}
}
