package bpe

import (
	"encoding/json"
	"unsafe"
)

// NOTE: the transport type is provided to provide easier structuring and destructuring into encoded types
//
// The common pattern: `t := *(*transport)(unsafe.Pointer(&p))` turns the Pair into a transport.
//
// It is OK to use unsafe here because `t` will be in a different memory location from p.

// transport is an internal type for IO stuff like JSON marshaling and the like.
type transport struct {
	Fst rune `json:"fst"`
	Snd rune `json:"snd"`
}

// MarshalJSON returns the JSON-encoded version of Pair
func (p Pair) MarshalJSON() ([]byte, error) {
	t := *(*transport)(unsafe.Pointer(&p))
	return json.Marshal(t)
}

// UnmarshalJSON unmarshals a JSON encoded Pair into the data structure itself
func (p *Pair) UnmarshalJSON(bs []byte) error {
	var t transport
	if err := json.Unmarshal(bs, &t); err != nil {
		return err
	}
	*p = *(*Pair)(unsafe.Pointer(&t))
	return nil
}
