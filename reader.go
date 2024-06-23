package avslz

import (
	"bufio"
	"encoding/binary"
	"io"
)

type reader struct {
	rd         io.Reader
	unread     [matchLimit * 8]byte
	head, tail int
	window     window
	code       [2]byte
	eof        bool
}

// NewReader returns an io.Reader that reads and decompresses data
// from rd.
func NewReader(rd io.Reader) io.Reader {
	if _, ok := rd.(io.ByteReader); !ok {
		rd = bufio.NewReader(rd)
	}
	return &reader{
		rd: rd,
	}
}

func (lz *reader) Read(p []byte) (n int, err error) {
	for len(p) > 0 {
		if lz.tail >= lz.head {
			if lz.eof {
				return n, io.EOF
			}

			if err = lz.inflate(); err != nil {
				if err == io.EOF {
					lz.eof = true
				} else {
					break
				}
			}
		}

		size := lz.head - lz.tail
		if lp := len(p); size > lp {
			size = lp
		}
		copy(p, lz.unread[lz.tail:lz.tail+size])

		lz.tail += size
		n += size
		p = p[size:]
	}

	return
}

func (lz *reader) put(b byte) {
	lz.window.put(b)
	lz.unread[lz.head] = b
	lz.head++
}

func (lz *reader) inflate() error {
	lz.head = 0
	lz.tail = 0

	flag, err := lz.rd.(io.ByteReader).ReadByte()
	if err != nil {
		return err
	}

	for i := 0; i < 8; i++ {
		if (flag>>i)&1 == 1 {
			b, err := lz.rd.(io.ByteReader).ReadByte()
			if err != nil {
				return err
			}
			lz.put(b)
		} else {
			if _, err := io.ReadFull(lz.rd, lz.code[:]); err != nil {
				return err
			}
			code := binary.BigEndian.Uint16(lz.code[:])
			if code == 0 {
				return io.EOF
			}

			offset := lz.window.head - int(code>>4)
			size := int(code)&0b1111 + matchThreshold
			for i := 0; i < size; i++ {
				lz.put(lz.window.at(offset + i))
			}
		}
	}

	return nil
}
