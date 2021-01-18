package bpe

// a funcMod is a struct holding all the modifier options for a function
type funcMod struct {
	buf []Pair // can be nil
}

// FuncOpt is an option to modify the behaviours of a function
type FuncOpt func(*funcMod)

// WithReuse uses the given (usually pre-allocated) buffer of Pairs
func WithReuse(buf []Pair) FuncOpt {
	return func(m *funcMod) {
		m.buf = buf
	}
}
