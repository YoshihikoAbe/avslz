package avslz

type window struct {
	head   int
	buffer [windowSize]byte
}

func (w *window) put(b byte) {
	w.buffer[w.head] = b
	w.head = (w.head + 1) % windowSize
}

func (w *window) at(i int) byte {
	return w.buffer[uint(i)%windowSize]
}
