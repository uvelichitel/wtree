package wtarray

import (
	"errors"
	"github.com/uvelichitel/wtree/bitmap64"
	"github.com/uvelichitel/wtree"
)

type WTArray struct {
	Count uint
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

type Maps []bitmap64.Bitmap64

//func (wt *WTArray)Edit(sym string, pos uint, d Dict) error{
//	if
//
//	return nil
//}
//func (wt WTArray) Count() uint {
//	return wt.Count
//}
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

func (wta WTArray) AccessDict(pos uint) string {
	wt, pos := wtree.Access(wta, pos)
	wta = wt.(WTArray)
	bit := (*wta.Maps)[wta.Mark].Get(pos)
	return (*wta.Dict)[2*wta.Mark+1+int(bit)-len(*wta.Maps)]
}

func (wt WTArray) TrackDict(sym string, count uint) (uint, error) {
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
	pos := wt.Count
	wt.Count++
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
		maps[k] = make(bitmap64.Bitmap64, 0)
	}
	wt.Maps = &maps
	return wt
}
