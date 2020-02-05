package bitmap64

import (
	"math/bits"
)

type Bitmap64 []uint64

func Mask(u uint64) uint64 {
	u |= u >> 1
	u |= u >> 2
	u |= u >> 4
	u |= u >> 8
	u |= u >> 16
	u |= u >> 32
	return u
}

func (bm Bitmap64) Lenth() int {
	return 64*len(bm) - bits.LeadingZeros64(bm[len(bm)-1]) - 1
}
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
func (bm *Bitmap64) Push(bit int8) {
	var U uint64
	if len(*bm) == 0 {
		*bm = append(*bm, 1)
	}
	m := (*bm)[len(*bm)-1]
	ma := Mask(m)
	if ma == ^U {
		(*bm)[len(*bm)-1] = m & (ma >> ((bit + 1) & 1))
		*bm = append(*bm, 1)
		return
	}
	m = m & (ma >> ((bit + 1) & 1))
	m = m + ma + 1
	(*bm)[len(*bm)-1] = m
	return
}
func (bm Bitmap64) Get(pos uint) int8 {
	return int8((bm[int(pos/64)] >> (pos % 64)) & 1)
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
	num++
	for {
		d = uint(bits.OnesCount64(bm[n]))
		if d == num {
			pos += 63
			return pos
		}
		if d > num {
			break
		}
		num -= d
		pos += 64
		n++
	}
	c = bm[n]
	c &= (Mask(c) >> 1)
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
	for ; num != 1; num-- {
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
	c &= (Mask(c) >> 1)
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
