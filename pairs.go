package bpe

import (
	"fmt"
)

// "const"
var negsymb = []byte("-")

// Pair is a pair of runes - it is an immutable tuple. Use P() to create a new Pair
type Pair struct{ fst, snd rune }

// P constructs a new Pair
func P(fst, snd rune) Pair { return Pair{fst, snd} }

// Fst returns the first projection
func (p Pair) Fst() rune { return p.fst }

// Snd returns the second projection
func (p Pair) Snd() rune { return p.snd }

// Eq is the comparison function for two pairs, p and q
func (p Pair) Eq(q Pair) bool { return p.fst == q.fst && p.snd == q.snd }

// Format implements fmt.Formatter
func (p Pair) Format(s fmt.State, c rune) {
	if c == 'd' {
		fmt.Fprintf(s, "(%d %d)", p.fst, p.snd)
		return
	}
	format := "%c"
	if c == 'q' {
		format = "%q"
	}
	s.Write([]byte("("))
	fmt.Fprintf(s, format, p.fst)
	s.Write([]byte(" "))
	snd := p.snd
	if snd < 0 {
		s.Write(negsymb)
		snd = -snd
	}
	fmt.Fprintf(s, format, snd)
	s.Write([]byte(")"))
}

func pairs(word string, buf []Pair) []Pair {
	buf = buf[:0]
	var prev rune
	for i, r := range word {
		if i == 0 {
			prev = r
			continue
		}
		p := Pair{prev, r}
		buf = append(buf, p)
		prev = r
	}
	return buf
}

// pairs2 is exactly the same as pairs,except operating on a word that is a []rune
func pairs2(word []rune, buf []Pair) []Pair {
	buf = buf[:0]
	var prev rune
	for i, r := range word {
		if i == 0 {
			prev = r
			continue
		}
		p := Pair{prev, r}
		buf = append(buf, p)
		prev = r
	}
	return buf
}

func indicesOf(p Pair, in []Pair) []int {
	retVal := make([]int, 0, 2) // a pair is unlikely to occur more than twice in a word
	for i, q := range in {
		if p.Eq(q) {
			retVal = append(retVal, i)
		}
	}
	return retVal
}

type replacedWord struct {
	id       int
	original string
}
