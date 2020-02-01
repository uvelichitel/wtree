package wtarray

import (
	"github.com/uvelichitel/wtree"
	"github.com/uvelichitel/wtree/bitmap64"
)


type Maps []bitmap64.Bitmap64

type WTArray struct {
	*Maps
	Mark int
}
func New(c int) WTArray {
	var wt WTArray
	if (c & (c - 1)) != 0 {
		//c--
		c |= c >> 1
		c |= c >> 2
		c |= c >> 4
		c |= c >> 8
		c |= c >> 16
		c++
	}
	maps := make(Maps, c-1, c-1)
	for k, _ := range maps {
		maps[k] = make(bitmap64.Bitmap64, 0)
	}
	wt.Maps = &maps
	return wt
}
func (wt WTArray) BitMap() wtree.Bitmap {
	return (*wt.Maps)[wt.Mark]
}
func (wt WTArray) RChild() wtree.WTree {
	wt.Mark = 2*wt.Mark + 2
	return wt
}
func (wt WTArray) LChild() wtree.WTree {
	wt.Mark = 2*wt.Mark + 1
	return wt
}
func (wt WTArray) Parrent() wtree.WTree {
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
func (wt *WTArray) ToLeaf(ind int) int8 {
	bit := int8(ind & 1)
	wt.Mark = (ind + len(*wt.Maps) - 1) / 2
	return bit
}
func (wt WTArray) FromLeaf(pos uint) int {
	bit := (*wt.Maps)[wt.Mark].Get(pos)
	return 2*wt.Mark + 1 + int(bit) - len(*wt.Maps)
}

func (wta WTArray) Access(pos uint) int {
	wt, pos := wtree.Access(wta, pos)
	wta = wt.(WTArray)
	bit := (*wta.Maps)[wta.Mark].Get(pos)
	return 2*wta.Mark + 1 + int(bit) - len(*wta.Maps)
}
func (wta WTArray) Track(ind int, count uint) uint {
	bit := int8(ind & 1)
	wta.Mark = (ind + len(*wta.Maps) - 1) / 2
	_, pos := wtree.Track(wta, count, bit)
	return pos
}
func (wt WTArray) Append(ind int) {
	b := wt.ToLeaf(ind)
	for {
		(*wt.Maps)[wt.Mark].Append(b)
		if wt.IsHead() {
			return
		}
		b = int8((wt.Mark - 1) & 1)
		wt = wt.Parrent().(WTArray)
	}
}

