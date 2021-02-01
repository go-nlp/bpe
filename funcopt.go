package bpe

// a funcMod is a struct holding all the modifier options for a function
type funcMod struct {
	buf     []Pair // can be nil
	markEOW bool   // mark end of word?
}

// FuncOpt is an option to modify the behaviours of a function
type FuncOpt func(*funcMod)

// WithReuse uses the given (usually pre-allocated) buffer of Pairs
func WithReuse(buf []Pair) FuncOpt {
	return func(m *funcMod) {
		m.buf = buf
	}
}

// MarkEOW is a modifier to inform the Learn function whether the end of the word should be marked.
func MarkEOW(t bool) FuncOpt {
	return func(m *funcMod) {
		m.markEOW = t
	}
}
