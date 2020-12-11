package udf

import "fmt"

type EntityID struct {
	Flags            uint8
	Identifier       [23]byte
	IdentifierSuffix [8]byte
}

func (e *EntityID) Show(name string) {
	fmt.Printf("%s: %d - %s - %v\n", name, e.Flags, string(e.Identifier[:]), e.IdentifierSuffix)
}

func NewEntityID(b []byte) EntityID {
	e := EntityID{Flags: b[0]}
	copy(e.Identifier[:], b[1:24])
	copy(e.IdentifierSuffix[:], b[24:32])
	return e
}
