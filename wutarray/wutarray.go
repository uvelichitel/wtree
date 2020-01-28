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

func (wta UWTArray) AccessDict(pos uint) string {
	wt, pos := wtree.Access(wta, pos)
	wta = wt.(UWTArray)
	bit := (*wta.Maps)[wta.Mark].Get(pos)
	return (*wta.Dict)[2*wta.Mark+1+int(bit)-len(*wta.Maps)]
}

func (wt UWTArray) TrackDict(sym string, count uint) (uint, error) {
	ind, err := wt.Dict.Lookup(sym)
	var pos uint
	if err != nil {
		return 0, errors.New("Symbol is absent in dictionary")
	}
	bit := int8(ind & 1)
	m := (ind + len(*wt.Maps) - 1) / 2
	wt.Mark = m
	_, pos = wtree.Track(wt, count, bit)
	return pos, nil
}
func (wt *UWTArray) Append(sym string) error {
	index, err := wt.Dict.Lookup(sym)
	if err != nil {
		index = len(*wt.Dict)
		if index < cap(*wt.Dict) {
			*wt.Dict = append(*wt.Dict, sym)
		} else {
			return errors.New("Not enough space")
		}
	}
	pos := wt.Count
	wt.Count++
	mark := wt.Layout.Select0(index)

}
type Layout bitmap64.Bitmap64

type Maps []bitmap64.Bitmap64

type Dict []string

func (d Dict) Lookup(s string) (int, error) {
	for k, v := range d {
		if v == s {
			return k, nil
		}
	}
	return 0, errors.New("Term not found in dictionary")
}
