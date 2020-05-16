package wtree

//Standard methods for bitmaps and succinct data structures.
type Bitmap interface {
	Rank1(uint) uint
	Rank0(uint) uint
	Select1(uint) uint
	Select0(uint) uint
	Get(uint) int8
}

//Standard methods for tree.
type WTree interface {
	BitMap() Bitmap
	RChild() WTree
	LChild() WTree
	Parrent() WTree
	IsLeaf() bool
	IsHead() bool
	IsLChild() bool
}

func Access(wt WTree, pos uint) (WTree, uint) {
	var bit int8
	for !wt.IsLeaf() {
		bit = wt.BitMap().Get(pos)
		if bit == 0 {
			pos = wt.BitMap().Rank0(pos) - 1
			wt = wt.LChild()
		} else {
			pos = wt.BitMap().Rank1(pos) - 1
			wt = wt.RChild()
		}
	}
	return wt, pos
}

func Track(wt WTree, count uint, bit int8) (WTree, uint) {
	lch := bit == 0
	for {
		if lch {
			count = wt.BitMap().Select0(count)
		} else {
			count = wt.BitMap().Select1(count)
		}
		if wt.IsHead() {
			break
		}
		lch = wt.IsLChild()
		wt = wt.Parrent()
	}
	return wt, count
}
