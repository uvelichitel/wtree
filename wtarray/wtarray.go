package wtarray

import (
	"errors"
	"math/bits"
)

type WTArray struct {
	Len uint
	*Maps
	Mark int
	*Dict
}
type Dict []string

func (d Dict) Lookup(s string) (int, error) {
	for k, v := range d {
		if v == s {
			return k, nil
		}
	}
	return 0, errors.New("Term not found in dictionary")
}

type Maps []Bitmap64

//func (wt *WTArray)Edit(sym string, pos uint, d Dict) error{
//	if
//
//	return nil
//}
func (wt WTArray) Count() uint {
	return wt.Len
}

func (wt WTArray) BitMap() Bitmap {
	return (*wt.Maps)[wt.Mark]
}
func (wt WTArray) RChild() WTree {
	wt.Mark = 2*wt.Mark + 2
	return wt
}
func (wt WTArray) LChild() WTree {
	wt.Mark = 2*wt.Mark + 1
	return wt
}
func (wt WTArray) Parrent() WTree {
	wt.Mark = (wt.Mark - 1) / 2
	return wt
}
func (wt WTArray) IsLeaf() bool {
	if wt.Mark >= len(*wt.Maps)/2 {
		return true
	} else {
		return false
	}
}
func (wt WTArray) IsLChild() bool {
	return 1&wt.Mark != 0
}
func (wt WTArray) IsHead() bool {
	if wt.Mark == 0 {
		return true
	} else {
		return false
	}
}
func (wt *WTArray) DictToLeaf(sym string) (int8, error) {
	ind, err := wt.Dict.Lookup(sym)
	if err != nil {
		return 0, err
	}
	bit := int8(ind & 1)
	m := (ind + len(*wt.Maps) - 1) / 2
	wt.Mark = m
	return bit, nil
}
func (wt WTArray) LeafToDict(pos uint) string {
	bit := (*wt.Maps)[wt.Mark].Get(pos)
	return (*wt.Dict)[2*wt.Mark+1+int(bit)-len(*wt.Maps)]
}

func (wt *WTArray) Append(sym string) error {
	var l, r, h int
	index, err := wt.Dict.Lookup(sym)
	if err != nil {
		index = len(*wt.Dict)
		if index < cap(*wt.Dict) {
			*wt.Dict = append(*wt.Dict, sym)
		} else {
			return errors.New("Not enough space")
		}
	}
	pos := wt.Len
	wt.Len++
	h = cap(*wt.Dict)
	l = 0
	r = h - 1
	for mark := 0; mark < len(*wt.Maps); {
		if index > (r+l)/2 {
			l = l + h/2
			h = h / 2
			(*wt.Maps)[mark].Set(1, pos)
			pos = (*wt.Maps)[mark].Rank1(pos) - 1
			mark = 2*mark + 2
		} else {
			r = r - h/2
			h = h / 2
			(*wt.Maps)[mark].Set(0, pos)
			pos = (*wt.Maps)[mark].Rank0(pos) - 1
			mark = 2*mark + 1
		}
	}
	return nil
}

func FromDictionary(d Dict) WTArray {
	var wt WTArray
	l := len(d)
	c := cap(d)
	if (c & (c - 1)) != 0 {
		c--
		c |= c >> 1
		c |= c >> 2
		c |= c >> 4
		c |= c >> 8
		c |= c >> 16
		c |= c >> 32
		c++
		a := make(Dict, l, c)
		copy(a, d)
		wt.Dict = &a
	} else {
		wt.Dict = &d
	}
	maps := make(Maps, c-1, c-1)
	for k, _ := range maps {
		maps[k] = make(Bitmap64, 0)
	}
	wt.Maps = &maps
	return wt
}

type Bitmap64 []uint64

func (bm *Bitmap64) Set(bit int8, pos uint) {
	n := int(pos / 64)
	d := n - len(*bm) + 1
	if d > 0 {
		a := make([]uint64, d)
		*bm = append(*bm, a...)
	}
	if bit == 0 {
		(*bm)[int(pos/64)] &^= 1 << (pos % 64)
	} else {
		(*bm)[n] |= 1 << (pos % 64)
	}
}

func (bm Bitmap64) Get(pos uint) int8 {
	return int8(bm[int(pos/64)] >> (pos % 64) & 1)
}

func (bm Bitmap64) Rank1(pos uint) (count uint) {
	var n uint
	for ; n < pos/64; n++ {
		count += uint(bits.OnesCount64(bm[int(n)]))
	}
	count += uint(bits.OnesCount64(bm[int(n)] << (63 - pos%64)))
	return
}
func (bm Bitmap64) Rank0(pos uint) (count uint) {
	count = pos - bm.Rank1(pos) + 1
	return
}

func (bm Bitmap64) Select1(num uint) (pos uint) {
	var c uint64
	var c1 uint32
	var c2 uint16
	var c3 uint8
	var n int
	var d uint
	for {
		d = uint(bits.OnesCount64(bm[n]))
		if d >= num {
			break
		}
		num -= d
		pos += 64
		n++
	}
	c = bm[n]
	c1 = uint32(c)
	d = uint(bits.OnesCount32(c1))
	if d < num {
		num -= d
		pos += 32
		c1 = uint32(c >> 32)
	}
	c2 = uint16(c1)
	d = uint(bits.OnesCount16(c2))
	if d < num {
		num -= d
		pos += 16
		c2 = uint16(c1 >> 16)
	}
	c3 = uint8(c2)
	d = uint(bits.OnesCount8(c3))
	if d < num {
		num -= d
		pos += 8
		c3 = uint8(c2 >> 8)
	}
	for ; num != 0; num-- {
		c3 &= c3 - 1
	}
	pos += uint(bits.TrailingZeros8(c3))
	return pos
}

func (bm Bitmap64) Select0(num uint) (pos uint) {
	var c uint64
	var c1 uint32
	var c2 uint16
	var c3 uint8
	var n int
	var d uint
	num++
	for {
		d = 64 - uint(bits.OnesCount64(bm[n]))
		if d >= num {
			break
		}
		num -= d
		pos += 64
		n++
	}
	c = bm[n]
	c1 = uint32(c)
	d = 32 - uint(bits.OnesCount32(c1))
	if d < num {
		num -= d
		pos += 32
		c1 = uint32(c >> 32)
	}
	c2 = uint16(c1)
	d = 16 - uint(bits.OnesCount16(c2))
	if d < num {
		num -= d
		pos += 16
		c2 = uint16(c1 >> 16)
	}
	c3 = uint8(c2)
	d = 8 - uint(bits.OnesCount8(c3))
	if d < num {
		num -= d
		pos += 8
		c3 = uint8(c2 >> 8)
	}
	for c3 = ^c3; num != 1; num-- {
		c3 &= c3 - 1
	}
	pos += uint(bits.TrailingZeros8(c3))
	return pos
}
