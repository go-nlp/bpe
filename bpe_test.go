package bpe

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"testing"
	"unsafe"

	"github.com/go-nlp/corpus"
	"github.com/stretchr/testify/assert"
)

type f struct{ bm map[rune]rune }

func (f f) Format(s fmt.State, c rune) {
	sortedKeys := make([]int, 0, len(f.bm))
	for k := range f.bm {
		sortedKeys = append(sortedKeys, int(k))
	}
	sort.Ints(sortedKeys)

	for _, k := range sortedKeys {
		fmt.Fprintf(s, "(%v, %c)\n", k, f.bm[rune(k)])
	}
}
func TestPreBPE(t *testing.T) {
	bm := PreBPE()
	t.Logf("%d", len(bm))
	t.Logf("%v", f{bm})
}

func Test_replaceInString(t *testing.T) {
	a := "Hello Worlld oll"
	p := Pair{'l', 'l'}

	ret := replaceInString(a, p, 'l')
	if ret != "Helo World ol" {
		t.Error("Expected something else")
	}

	b := "no replaceable pairs of characters"
	ret = replaceInString(b, p, 'l')
	if ret != b {
		t.Errorf("No change should be made. Got %q", ret)
	}

	c := "llllllllll"
	ret = replaceInString(c, p, 'X')
	if ret != "XXXXX" {
		t.Errorf("Got %q", ret)
	}
}

func TestPairs(t *testing.T) {
	assert := assert.New(t)
	a := "Hello"
	ps := []Pair{{'H', 'e'}, {'e', 'l'}, {'l', 'l'}, {'l', 'o'}}
	ret := Pairs(a)
	assert.Equal(ps, ret, "Expected %v. Got %v", ps, ret)

	// with buffer
	buf := make([]Pair, 10)
	ret = Pairs(a, WithReuse(buf))
	assert.Equal(buf[:4], ret)
	assert.Equal(ps, buf[:4])

	b := "e" // edge case
	ret = Pairs(b)
	ps = []Pair{}
	for _, p := range ret {
		t.Logf("%q %q", p.fst, p.snd)
	}
	assert.Equal(ps, ret, "Expected %c. Got %c", ps, ret)
	assert.Equal(fmt.Sprintf("%v", ps), fmt.Sprintf("%v", ret))
}

func TestPairsRunes(t *testing.T) {
	assert := assert.New(t)
	a := []rune("Hello")
	ps := []Pair{{'H', 'e'}, {'e', 'l'}, {'l', 'l'}, {'l', 'o'}}
	ret := PairsRunes(a)
	assert.Equal(ps, ret, "Expected %v. Got %v", ps, ret)

	// with buffer
	buf := make([]Pair, 10)
	ret = PairsRunes(a, WithReuse(buf))
	assert.Equal(buf[:4], ret)
	assert.Equal(ps, buf[:4])

	b := []rune("e") // edge case
	ret = PairsRunes(b)
	ps = []Pair{}
	for _, p := range ret {
		t.Logf("%q %q", p.fst, p.snd)
	}
	assert.Equal(ps, ret, "Expected %c. Got %c", ps, ret)
	assert.Equal(fmt.Sprintf("%v", ps), fmt.Sprintf("%v", ret))
}

func Test_updateStats(t *testing.T) {
	const text = `hello world el melodies`
	const text2 = `hxlo world el mxodies`
	assert := assert.New(t)
	c, _ := corpus.Construct(corpus.WithWords(SimpleTokenizer(text)))
	s := PairStats(c)
	p := Pair{'e', 'l'}
	r := replacePair(s, p)
	updateStats(&s, r, p)
	tidyStats(s.Stats)
	tidyIndices(s.Indices)
	assert.Equal('x', s.MaxRune)

	d, _ := corpus.Construct(corpus.WithWords(SimpleTokenizer(text2)))
	S := PairStats(d)
	assert.Equal(S.Stats, s.Stats)
	assert.Equal(S.Indices, s.Indices)
}

func TestLearn(t *testing.T) {
	assert := assert.New(t)
	bs, err := ioutil.ReadFile("testdata/corpus.txt")
	if err != nil {
		t.Fatalf("Unable to proceed. Error while reading external corpus: %v", err)
	}
	text := *(*string)(unsafe.Pointer(&bs))
	c, _ := corpus.Construct(corpus.WithWords(SimpleTokenizer(text)))

	ps, err := Learn(c, 10, 2, true)
	if err != nil {
		t.Error(err)
	}

	correct := []Pair{
		{'t', 'h'},
		{'h', -'e'},
		{'i', 'n'},
		{'a', 'n'},
		{'h', 'e'},
		{'o', 'u'},
		{'e', 'r'},
		{'n', -'d'},
		{'h', 'a'},
		{'i', 't'},
	}
	correctRepl := map[Pair]int32{
		{'a', 'n'}:  8225,
		{'e', 'r'}:  8228,
		{'h', -'e'}: 8223,
		{'h', 'a'}:  8230,
		{'h', 'e'}:  8226,
		{'i', 'n'}:  8224,
		{'i', 't'}:  8231,
		{'n', -'d'}: 8229,
		{'o', 'u'}:  8227,
		{'t', 'h'}:  8222,
	}

	assert.Equal(correct, ps.Pairs)
	assert.Equal(c, ps.Corpus)
	assert.Equal(correctRepl, ps.Replacements)
	assert.Equal('‧', ps.MaxRune)
}

func TestFormat(t *testing.T) {
	pairs := []Pair{
		{'t', 'h'},
		{'h', -'e'}, // when a character is next to a space it gets encoded to be negative
	}
	asNum := "[(116 104) (104 -101)]"
	asQuo := "[('t' 'h') ('h' -'e')]"
	byDefault := "[(t h) (h -e)]"
	assert.Equal(t, asNum, fmt.Sprintf("%d", pairs))
	assert.Equal(t, asQuo, fmt.Sprintf("%q", pairs))
	assert.Equal(t, byDefault, fmt.Sprintf("%v", pairs))
}

func TestJSON(t *testing.T) {
	ps := []Pair{
		{'a', 'b'},
		{'c', 'd'},
		{'e', '\n'},
	}
	bs, err := json.Marshal(ps)
	if err != nil {
		t.Fatal(err)
	}
	var ps2 []Pair
	if err = json.Unmarshal(bs, &ps2); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, ps, ps2)
}

func benchLearnSetup() *corpus.Corpus {
	bs, err := ioutil.ReadFile("testdata/corpus.txt")
	if err != nil {
		panic(err)
	}
	text := *(*string)(unsafe.Pointer(&bs))
	c, _ := corpus.Construct(corpus.WithWords(SimpleTokenizer(text)))
	return c
}

func BenchmarkWithBuf(b *testing.B) {
	b.StopTimer()
	c := benchLearnSetup()
	b.ResetTimer()
	b.StartTimer()

	var ps Encoder
	for i := 0; i < b.N; i++ {
		ps, _ = Learn(c, 100, 2, true)
	}
	_ = ps
}

func BenchmarkWithoutBuf(b *testing.B) {
	b.StopTimer()
	c := benchLearnSetup()
	b.ResetTimer()
	b.StartTimer()

	var ps Encoder
	for i := 0; i < b.N; i++ {
		ps, _ = Learn(c, 100, 2, true)
	}
	_ = ps
}
