package avslz

import (
	"io"
)

// A Writer is an io.Writer that compresses the data written to it.
type Writer struct {
	wr         io.Writer
	unread     [1024]byte
	unreadHead int
	// accommodates the largest possible payload size
	output          [1 + 2*8]byte
	outputHead      int
	window          window
	flag, flagShift uint8
}

// NewWriter returns a new Writer that uses wr as its underlying io.Writer.
func NewWriter(wr io.Writer) *Writer {
	return &Writer{
		wr:         wr,
		outputHead: 1,
	}
}

// Write compresses p and writes it to the underlying io.Writer.
// It should be assumed that the compressed data is not fully
// written until the Writer has been closed.
func (lz *Writer) Write(p []byte) (n int, err error) {
	for len(p) > 0 {
		if lz.unreadHead >= len(lz.unread) {
			if err = lz.deflate(); err != nil {
				return
			}
			lz.unreadHead = 0
		}

		size := len(lz.unread) - lz.unreadHead
		if lp := len(p); size > lp {
			size = lp
		}
		copy(lz.unread[lz.unreadHead:], p)

		lz.unreadHead += size
		n += size
		p = p[size:]
	}
	return
}

// Close flushes any unwritten compressed data, and writes
// the end-of-file marker to the underlying io.Writer.
func (lz *Writer) Close() (err error) {
	if err = lz.deflate(); err != nil {
		return
	}
	if err = lz.flush(); err != nil {
		return
	}
	_, err = lz.wr.Write([]byte{0, 0, 0})
	return
}

func (lz *Writer) flush() (err error) {
	lz.output[0] = lz.flag
	_, err = lz.wr.Write(lz.output[:lz.outputHead])
	lz.flagShift = 0
	lz.flag = 0
	lz.outputHead = 1
	return
}

func (lz *Writer) findMatch(tail int) (bo, bs int) {
	if tail+matchThreshold >= lz.unreadHead {
		return
	}

	for i := 0; i < searchDistance; i++ {
		size := 0

		for ; size+tail < lz.unreadHead; size++ {
			wo := lz.window.head - i + size
			if (wo >= lz.window.head) ||
				(lz.window.at(wo) != lz.unread[tail+size]) {
				break
			}

			if size >= matchLimit {
				return i, size
			}
		}

		if size > bs {
			bs = size
			bo = i
		}
	}

	return
}

func (lz *Writer) deflate() error {
	for i := 0; i < lz.unreadHead; {
		if lz.flagShift >= 8 {
			if err := lz.flush(); err != nil {
				return err
			}
		}

		offset, size := lz.findMatch(i)
		if size >= matchThreshold {
			lz.output[lz.outputHead] = byte(offset >> 4)
			lz.outputHead++
			lz.output[lz.outputHead] = byte((offset&0b1111)<<4 | (size - matchThreshold))
			lz.outputHead++

			for k := 0; k < size; k++ {
				lz.window.put(lz.unread[i])
				i++
			}
		} else {
			lz.flag |= 1 << lz.flagShift

			b := lz.unread[i]
			i++
			lz.output[lz.outputHead] = b
			lz.outputHead++
			lz.window.put(b)
		}

		lz.flagShift++
	}

	return nil
}
