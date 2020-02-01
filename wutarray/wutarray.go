package wutarray

import (
	"github.com/uvelichitel/wtree"
	"github.com/uvelichitel/wtree/bitmap64"
)

type Maps []bitmap64.Bitmap64

type UWTArray struct {
	*Maps
	Mark   int
	Layout *bitmap64.Bitmap64
}

//func FromLayout()

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
		(*wta.Maps)[wta.Mark].Append(b)
		b = int8((wta.Mark - 1) & 1)
		wta = wta.Parrent().(UWTArray)
		if wta.IsHead() {
			(*wta.Maps)[wta.Mark].Append(b)
			return
		}
	}
}
