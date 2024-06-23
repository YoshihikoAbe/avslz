package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/YoshihikoAbe/avslz"
)

func main() {
	var (
		decompress   bool
		ipath, opath string

		rd  io.Reader      = os.Stdin
		wr  io.WriteCloser = os.Stdout
		err error
	)

	flag.BoolVar(&decompress, "d", false, "Enable decompression")
	flag.StringVar(&ipath, "i", "-", "Path to the input file, or \"-\" for standard input")
	flag.StringVar(&opath, "o", "-", "Path to the output file, or \"-\" for standard output")
	flag.Parse()

	if ipath != "-" {
		if rd, err = os.Open(ipath); err != nil {
			fatal(err)
		}
	}

	if opath != "-" {
		if wr, err = os.Create(opath); err != nil {
			fatal(err)
		}
	}

	if decompress {
		rd = avslz.NewReader(rd)
	} else {
		wr = avslz.NewWriter(wr)
	}

	if _, err := io.Copy(wr, rd); err != nil {
		fatal("failed to compress/decompress data:", err)
	}
	if !decompress {
		if err := wr.Close(); err != nil {
			fatal("failed to close output:", err)
		}
	}
}

func fatal(v ...any) {
	fmt.Println(v...)
	os.Exit(1)
}
