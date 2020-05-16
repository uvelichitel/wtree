package indexer

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"strconv"
	"sync"

	"github.com/uvelichitel/wtree/dict"
	"github.com/uvelichitel/wtree/hwtree"
)

//Indexer build from wavelet-tree and dictionary of terms.
type HWT struct {
	hwtree.HWTArray
	dict.Dict
}

func New(s []string, f []uint) (h HWT) { //TODO check errors
	l, t := hwtree.Huffman(f)
	d := make([]string, len(t))
	for k, v := range t {
		d[k] = s[v]
	}
	h.HWTArray = hwtree.New(l)
	h.Dict = dict.New(d)
	return
}

//Read frequency list in standart format
//"term" "frequency"
//each line
func FromFreqList(r io.Reader) (h HWT, err error) {
	rs := csv.NewReader(r)
	s := make([]string, 0)
	f := make([]uint, 0)
	var ff int
	var rec []string
	for {
		rec, err = rs.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return h, err
		}
		s = append(s, rec[0])
		ff, err = strconv.Atoi(rec[1])
		if err != nil {
			return h, err
		}
		f = append(f, uint(ff))
	}
	h = New(s, f)
	return h, nil
}

//Common writer interface implementation for indexer
func (wt HWT) Write(p []byte) (n int, err error) {
	s := bytes.Fields(p)
	var ind int
	for _, v := range s {
		ind, err = wt.Dict.Lookup(string(v))
		if err == nil {
			n += len(v)
			wt.Append(ind)
		} else {
			return n, err
		}
	}
	return n, nil
}

//Reader interface implementation.
func (wt HWT) Consume(r io.Reader) (err error) {
	b := bufio.NewScanner(r)
	b.Split(bufio.ScanWords)
	var t string
	var ind int
	for b.Scan() {
		t = b.Text()
		ind, err = wt.Dict.Lookup(string(t))
		if err != nil {
			wt.Append(ind)
		} else {
			return err
		}
	}
	return b.Err()
}

//Classic wavelet-tree access metod implementation.
func (wt HWT) Access(pos uint) (string, error) {
	if pos >= wt.Maps[0].Length() {
		return "", errors.New("Position out of range ")
	}
	return wt.Dict.Get(wt.HWTArray.Access(pos)), nil
}

//Multithreaded access to position range
func (wt HWT) AccessRange(from, to uint) (s []string, err error) {
	if to >= wt.Maps[0].Length() {
		return s, errors.New("Position out of range ")
	}
	s = make([]string, int(to-from+1))
	var wg sync.WaitGroup
	for i := uint(0); i <= to-from; i++ {
		wg.Add(1)
		go func(i uint) {
			s[i] = wt.Dict.Get(wt.HWTArray.Access(from + i))
			wg.Done()
		}(i)
	}
	wg.Wait()
	return s, nil
}

//Track metod of data structure wavelet-tree. Return position of counted c inclusion of term s in incoming sequence.
func (wt HWT) Track(s string, c uint) (uint, error) {
	ind, err := wt.Dict.Lookup(s)
	if err != nil {
		return 0, err
	}
	m := int(wt.Layout.Select0(uint(ind) / 2))
	lch := (ind + 1) & 1
	l := uint(wt.Maps[m].Length())
	if (lch != 0 && (wt.Maps[m].Rank0(l-1) < c)) || (lch == 0 && (wt.Maps[m].Rank1(l-1) < c)) {
		return 0, errors.New("Too much to count")
	}
	return wt.HWTArray.Track(ind, c), nil
}

//Multithreaded track. Return array of term s inclusions in sequence.
func (wt HWT) TrackAll(s string) (r []uint, err error) {
	ind, err := wt.Dict.Lookup(s)
	if err != nil {
		return r, err
	}
	var c uint
	m := int(wt.Layout.Select0(uint(ind) / 2))
	lch := (ind + 1) & 1
	l := wt.Maps[m].Length()
	if lch != 0 {
		c = wt.Maps[m].Rank0(l - 1) -1 
	} else {
		c = wt.Maps[m].Rank1(l - 1) -1
	}
	r = make([]uint, c)
	var wg sync.WaitGroup
	for i := uint(0); i < c; i++ {
		wg.Add(1)
		go func(i uint) {
			r[i] = wt.HWTArray.Track(ind, i)
			wg.Done()
		}(i)
	}
	wg.Wait()
	return
}
