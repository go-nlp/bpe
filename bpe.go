// bpe provides a byte-pair encoding for text
package bpe

// PreBPE is a function that provides mapping for runes. This function is used for handling large text corpuses,
// and it is derived from OpenAI's GPT-2.
// The original code may be found here: https://github.com/openai/gpt-2/blob/master/src/encoder.py
//
// The original comments from GPT-2 clarifies:
// 	The reversible bpe codes work on unicode strings.
// 	This means you need a large # of unicode characters in your vocab if you want to avoid UNKs.
// 	When you're at something like a 10B token dataset you end up needing around 5K for decent coverage.
//	This is a signficant percentage of your normal, say, 32K bpe vocab.
// 	To avoid that, we want lookup tables between utf-8 bytes and unicode strings.
// 	And avoids mapping to whitespace/control characters the bpe code barfs on.
//
// It is unsure what utiltiy this provides now, given the design direction of the BPE package has gone in a slightly
// different direction - this package deals with runes, instead of messing around with strings and bytes.
// We sacrifice memory for readability and understandability.
func PreBPE() map[rune]rune {
	bs := make([]rune, 0, (127-33)+(173-161)+(256-174))
	for i := 33; i < 127; i++ {
		bs = append(bs, rune(i))
	}
	for i := 161; i < 173; i++ {
		bs = append(bs, rune(i))
	}
	for i := 174; i < 256; i++ {
		bs = append(bs, rune(i))
	}

	var n int
	cs := make([]rune, len(bs))
	copy(cs, bs)
	for i := 0; i < 256; i++ {
		if !inRange(rune(i), bs) {
			bs = append(bs, rune(i))
			cs = append(cs, rune(256+n))
			n++
		}
	}
	bytemap := make(map[rune]rune)
	for i, r := range bs {
		bytemap[r] = cs[i]
	}
	return bytemap
}

// Pairs returns the Pairs of runes found in a word (as string)
func Pairs(word string, opts ...FuncOpt) []Pair {
	var m funcMod
	for _, opt := range opts {
		opt(&m)
	}
	if m.buf == nil && len(word) > 0 {
		m.buf = make([]Pair, 0, len([]rune(word))-1)
	}

	return pairs(word, m.buf)
}

// PairsRunes returns the Pairs of runes found in a word (as []rune)
func PairsRunes(word []rune, opts ...FuncOpt) []Pair {
	var m funcMod
	for _, opt := range opts {
		opt(&m)
	}
	if m.buf == nil && len(word) > 0 {
		m.buf = make([]Pair, 0, len([]rune(word))-1)
	}

	return pairs2(word, m.buf)
}

// PairsWithReuse is the Pairs function, but with a buffer passed in specifically.
func PairsWithReuse(word string, buf []Pair) []Pair {
	return pairs(word, buf)
}

// PairsRunesWithReuse is the PairsRunes function, but with a buffer passed in specifically.
func PairsRunesWithReuse(word []rune, buf []Pair) []Pair {
	return pairs2(word, buf)
}

// UTIL

func inRange(r rune, rs []rune) bool {
	for _, s := range rs {
		if s == r {
			return true
		}
	}
	return false
}
