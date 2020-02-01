package dict

import(
	"errors"
)

type Dict []string

func (d Dict) Lookup(s string) (int, error) {
	for k, v := range d {
		if v == s {
			return k, nil
		}
	}
	return 0, errors.New("Term not found in dictionary")
}

