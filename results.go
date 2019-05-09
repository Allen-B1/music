package main

import (
	"strconv"
	"strings"
)

type Results struct {
	Composer bool
	Name     uint8
	Key      bool
}

func (r Results) Total() uint {
	total := uint(0)
	if r.Composer {
		total += 3
	}
	if r.Name == 1 {
		total += 3
	} else if r.Name >= 2 {
		total += 6
	}
	if r.Key {
		total += 1
	}
	return total
}

func (r Results) String() string {
	out := make([]string, 0, 3)
	if r.Composer {
		out = append(out, "c")
	}
	out = append(out, "n"+strconv.FormatUint(uint64(r.Name), 10))
	if r.Key {
		out = append(out, "k")
	}
	return strings.Join(out, ",")
}

func NewResults(in string) Results {
	out := Results{}
	ina := strings.Split(in, ",")
	for _, item := range ina {
		if len(item) == 0 {
			continue
		}
		switch item[0:1] {
		case "c":
			out.Composer = true
		case "k":
			out.Key = true
		case "n":
			if len(item) == 1 {
				continue
			}
			n, err := strconv.ParseUint(item[1:], 10, 8)
			if err == nil {
				out.Name = uint8(n)
			}
		}
	}
	return out
}

func NewResultsFromPiece(i *Incipit, composer string, name string, key string) Results {
	results := Results{}
	if strings.EqualFold(name, i.Name) {
		results.Name = 2
	} else if strings.Contains(strings.ToLower(i.Name), strings.ToLower(name)) {
		results.Name = 1
	}
	if strings.EqualFold(composer, i.Composer) {
		results.Composer = true
	}
	if strings.EqualFold(key, i.Key) {
		results.Key = true
	}
	return results
}
