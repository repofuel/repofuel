package sets

import "github.com/cheekybits/genny/generic"

//go:generate genny -in=set_generic.go -out=../engine/set_gnerated.go -pkg=engine gen "Item=string,Developer,Commit"
//go:generate genny -in=set_generic.go -out=../identifier/set_gnerated.go -pkg=identifier gen "Item=Hash"
//go:generate genny -in=set_generic.go -out=../manage/set_gnerated.go -pkg=manage gen "Item=*Queue"

type Item generic.Type

type ItemSet map[Item]struct{}

func NewItemSet() ItemSet {
	return make(ItemSet)
}

func (s ItemSet) Add(v Item) {
	s[v] = struct{}{}
}

func (s ItemSet) Update(s2 ItemSet) {
	for v := range s2 {
		s[v] = struct{}{}
	}
}

func (s ItemSet) Count() int {
	return len(s)
}

func (s ItemSet) Has(v Item) bool {
	_, ok := s[v]
	return ok
}

func (s ItemSet) Slice() []Item {
	slice := make([]Item, len(s))
	var i int
	for v := range s {
		slice[i] = v
		i += 1
	}
	return slice
}

func (s ItemSet) Delete(v Item) {
	delete(s, v)
}

func (s ItemSet) Copy() ItemSet {
	s2 := make(ItemSet, len(s))
	for v := range s {
		s2[v] = struct{}{}
	}
	return s2
}
