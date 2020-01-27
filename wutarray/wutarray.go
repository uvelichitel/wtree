package wutarray

import(
	"errors"
	"github.com/uvelichitel/wtree/bitmap64"
	"github.com/uvelichitel/wtree"
)

type UWTArray struct {
	Count uint
	*Maps
	Mark int
	*Dict
	*Layout
}

func (wt UWTArray) BitMap() wtree.Bitmap {
	return (*wt.Maps)[wt.Mark]
}
func (wt UWTArray) LChild() wtree.WTree {
	wt.Mark = wt.Layout.Rank1(wt.Mark -1)*2 + 1
	return wt
}
func (wt UWTArray) RChild() wtree.WTree {
	wt.Mark = wt.Layout.Rank1(wt.Mark -1)*2 + 2
	return wt
}
func (wt UWTArray) Parrent() wtree.WTRee{
	wt.Mark = wt.Layout.Select1((wt.Mark + 1)/2) - 1
	return wt
}
func (wt UWTArray) IsLChild() bool{
	return wt.Mark & 1 != 0
}
func (wt UWTArray) IsHead() bool{
	return wt.Mark == 0
}
func (wt UWTArray) IsLeaf() bool{
	return (*wt.Layout).Get(wt.Mark) == 0
}

type Layout bitmap64.Bitmap64

type Maps []bitmap64.Bitmap64

type Dict []string

