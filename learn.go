package bpe

import (
	"strings"

	"github.com/chewxy/lingo/corpus"
)

// Tokenizer is a function that tokenizes a string. This library provides a simple tokenizer.
type Tokenizer func(a string) []string

// SimpleTokenizer is a simple tokenizer of text
func SimpleTokenizer(a string) []string { return strings.Split(strings.Trim(a, "\r\n "), " ") }

// Statistics is the statistics of a corpus, used to figure out which pairs to replace.
type Statistics struct {
	Stats   map[Pair]int
	Indices map[Pair]map[int]int
	Corpus  *corpus.Corpus
	MaxRune rune
}

// PairStats returns the occurence frequencies of pairs of runes. It also construct an index of pairs to the word ID along its frequency
func PairStats(c *corpus.Corpus, opts ...FuncOpt) Statistics {
	stats := make(map[Pair]int)
	indices := make(map[Pair]map[int]int) // pair:{wordid:freq}
	// ENHANCEMENT: indices should have its own data struct
	var maxRune rune
	for i := 0; i < c.Size(); i++ {
		word, _ := c.Word(i)
		freq := c.WordFreq(word)

		ps := Pairs(word, opts...)
		for j, p := range ps {
			// for replacement rune
			if r := p.Fst(); r > maxRune {
				maxRune = r
			}
			if r := p.Snd(); r > maxRune {
				maxRune = r
			}
			if j == len(ps)-1 {
				p.snd = rune(-(int32(p.snd))) // the negative is a hack to mark the end of a word symbol
			}
			stats[p] += freq
			if indices[p] == nil {
				indices[p] = make(map[int]int)
			}
			indices[p][i]++
		}
	}
	return Statistics{
		Stats:   stats,
		Indices: indices,
		Corpus:  c,
		MaxRune: maxRune,
	}
}

// Encoder represents a state that may be used to encode a word
type Encoder struct {
	Corpus       *corpus.Corpus
	Pairs        []Pair
	Replacements map[Pair]rune
	MaxRune      rune
}

// Learn learns an Encoder from the given data in the corpus in the input.
func Learn(c *corpus.Corpus, symbols, minFreq int) (retVal Encoder, err error) {
	// if there are any preallocated []Pair that is being used, they will be safe for reuse once this function finishes
	stats := PairStats(c)

	var list []Pair
	rep := make(map[Pair]rune)
	for i := 0; i < symbols; i++ {
		m := mode(stats.Stats)

		// TODO: probably missing
		if stats.Stats[m] < minFreq {
			break // TODO error
		}

		replacements := replacePair(stats, m)
		updateStats(&stats, replacements, m)
		rep[m] = stats.MaxRune
		list = append(list, m)
		//log.Printf("mode %v replacements %v", m, replacements)
	}

	return Encoder{
		Corpus:       c,
		Pairs:        list,
		Replacements: rep,
		MaxRune:      stats.MaxRune,
	}, nil
}

// replacePair returns a list of replacements
func replacePair(stats Statistics, old Pair) (retVal []replacedWord) {
	c := stats.Corpus
	maxRune := stats.MaxRune
	indices := stats.Indices

	replacement := maxRune + 1
	retVal = make([]replacedWord, 0, len(indices[old]))
	for id, freq := range indices[old] {
		if freq < 1 {
			continue
		}
		word, _ := c.Word(id)

		newWord := replaceInString(word, old, replacement)
		c.ReplaceWord(id, newWord)
		retVal = append(retVal, replacedWord{id, word})
	}
	return retVal
}

// updateStats must be called immediately after replacePair
func updateStats(stats *Statistics, replacements []replacedWord, old Pair) {
	rr := stats.MaxRune + 1
	for _, r := range replacements {
		original := r.original

		ps := Pairs(original)
		is := indicesOf(old, ps)

		for _, i := range is {
			switch i {
			case -1:
				// error
			case 0:
				if len(ps) == 1 {
					continue
				}
				// replace next
				next := ps[i+1]
				p := P(rr, next.Snd())

				// update stats
				stats.Stats[next]--
				stats.Stats[p]++

				updateIndices(stats, next, p, r.id)
			case len(ps) - 1:
				// replace previous
				prev := ps[i-1]
				p := P(prev.Fst(), rr)

				//update stats
				stats.Stats[prev]--
				stats.Stats[p]++

				updateIndices(stats, prev, p, r.id)
			default:
				// replace previous and next
				prev := ps[i-1]
				p := P(prev.Fst(), rr)

				// update stats
				stats.Stats[prev]--
				stats.Stats[p]++

				updateIndices(stats, prev, p, r.id)

				next := ps[i+1]
				q := P(rr, next.Snd())

				// update stats
				stats.Stats[next]--
				stats.Stats[q]++

				updateIndices(stats, next, q, r.id)
			}
		}
	}
	delete(stats.Stats, old)
	delete(stats.Indices, old)
	stats.MaxRune++
}

func updateIndices(stats *Statistics, old, new Pair, wordID int) {
	// reduce the count of the old pair
	if _, ok := stats.Indices[old][wordID]; ok {
		stats.Indices[old][wordID]--
	}
	if stats.Indices[old][wordID] <= 0 {
		delete(stats.Indices[old], wordID)
	}

	// insert and update count of new pair
	if stats.Indices[new] == nil {
		stats.Indices[new] = make(map[int]int)
	}
	stats.Indices[new][wordID]++
}

// UTIL

func mode(a map[Pair]int) Pair {
	var maxFreq int = -1
	var max Pair
	for k, v := range a {
		// because Go's maps are nondeterministic,
		// we have to also compare the internals of a pair should there be a match to maxFreq.
		// This way we can always have deterministic results (makes testing easier)
		//
		// The choice of k < max is arbitrary
		if v > maxFreq {
			max = k
			maxFreq = v
		} else if v == maxFreq && (k.Fst() < max.Fst() || (k.Fst() == max.Fst() && k.Snd() < max.Snd())) {
			max = k
			maxFreq = v
		}
	}
	return max
}

func replaceInString(s string, p Pair, r rune) string {
	rs := []rune(s)
	fst := p.Fst()
	snd := p.Snd()
	for i := 0; i < len(rs); i++ {
		if rs[i] == fst && (i+1 < len(rs)) && rs[i+1] == snd {
			rs[i] = r
			if i+1 == len(rs)-1 {
				rs = rs[:i+1]
				break
			}
			copy(rs[i+1:], rs[i+2:])
			rs = rs[:len(rs)-1]
		}
	}
	return string(rs)
}

// tidyStats cleans up unused Pairs
func tidyStats(stats map[Pair]int) {
	var dels []Pair
	for k, v := range stats {
		if v <= 0 {
			dels = append(dels, k)
		}
	}
	for _, d := range dels {
		delete(stats, d)
	}
}

// tidyIndices cleans up unused Pairs from the indices
func tidyIndices(indices map[Pair]map[int]int) {
	var dels []Pair
	for k, v := range indices {
		if len(v) == 0 || v == nil {
			dels = append(dels, k)
		}
	}
	for _, d := range dels {
		delete(indices, d)
	}
}
