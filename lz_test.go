package avslz_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/YoshihikoAbe/avslz"
)

var (
	uncompressedTestCase = mustRead("testcases/image.bmp")
	compressedTestCase   = mustRead("testcases/image.bmp.lz")
)

func mustRead(filename string) []byte {
	b, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return b
}

func TestRoundTrip(t *testing.T) {
	compressed := &bytes.Buffer{}
	wr := avslz.NewWriter(compressed)
	if _, err := wr.Write(uncompressedTestCase); err != nil {
		t.Fatal(err)
	}
	if err := wr.Close(); err != nil {
		t.Fatal(err)
	}

	decompressed := &bytes.Buffer{}
	if _, err := io.Copy(decompressed, avslz.NewReader(compressed)); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(decompressed.Bytes(), uncompressedTestCase) {
		t.Fatal("round trip failed")
	}
}

func TestReader(t *testing.T) {
	got, err := io.ReadAll(avslz.NewReader(bytes.NewReader(compressedTestCase)))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, uncompressedTestCase) {
		t.Fatal("decompressed data does not match the original data")
	}
}

func BenchmarkWriter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		wr := avslz.NewWriter(io.Discard)
		if _, err := wr.Write(uncompressedTestCase); err != nil {
			b.Fatal(err)
		}
		if err := wr.Close(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rd := avslz.NewReader(bytes.NewReader(compressedTestCase))
		if _, err := io.Copy(io.Discard, rd); err != nil {
			b.Fatal(err)
		}
	}
}
