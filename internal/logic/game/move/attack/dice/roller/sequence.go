package roller

type Sequence struct {
	i   int
	seq []int
}

var _ Roller = (*Sequence)(nil)

func WithSequence(seq []int) Roller {
	return &Sequence{-1, seq}
}

// Roll returns the next number in the sequence.
func (s *Sequence) Roll() int {
	s.i = (s.i + 1) % len(s.seq)

	return s.seq[s.i]
}
