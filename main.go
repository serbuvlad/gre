package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"unicode/utf8"
)

type entry struct {
	r    rune
	size int
}

type reader struct {
	r   io.RuneReader
	buf []entry
	i   int
}

func (r *reader) ReadRune() (rune rune, size int, err error) {
	if r.i == -1 {
		rune, size, err = r.r.ReadRune()
		if err == nil {
			r.buf = append(r.buf, entry{r: rune, size: size})
		}
		return
	}

	if r.i == len(r.buf) {
		r.i = -1
		return r.ReadRune()
	}

	rune, size = r.buf[r.i].r, r.buf[r.i].size
	r.i++
	return
}

func init() {
	log.SetFlags(0)
	log.SetPrefix("gre: ")
}

var (
	bigl, bigh, h, l, o, s, v, wantc, wantn, yy bool
	x, y string
	re, xre *regexp.Regexp
	out = bufio.NewWriter(os.Stdout)

	ret = 1
)

func init() {
	flag.StringVar(&x, "x", ".*\n", "regular expression of section")
	flag.StringVar(&y, "y", "", "regular expression of section separator (overwrites x)")
	flag.BoolVar(&o, "o", false, "print only matching part (overwrites x, y, v and n)")

	flag.BoolVar(&v, "v", false, "converse; print non-matching section")

	flag.BoolVar(&h, "h", false, "supress head")
	flag.BoolVar(&bigh, "H", false, "force head")
	flag.BoolVar(&wantn, "n", false, "print section number")

	flag.BoolVar(&wantc, "c", false, "print only count of matching sections")

	flag.BoolVar(&l, "l", false, "print name of matching files")
	flag.BoolVar(&bigl, "L", false, "set l and v")

	flag.BoolVar(&s, "s", false, "print nothing. return only status")
}

func main() {
	flag.Parse()

	if bigl {
		l = true
		v = true
	}

	if o {
		wantn = false
	}

	if y != "" && !o {
		x = y
		yy = true
	}

	if flag.NArg() < 1 {
		log.Fatal("missing expression\n")
	}
	if flag.NArg() < 3 && !bigh {
		h = true
	}

	var err error

	re, err = regexp.Compile(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	if !o {
		xre, err = regexp.Compile(x)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		xre = re
	}

	switch flag.NArg() {
	case 1:
		xrep(os.Stdin, "stdin")
	default:
		for i := 1; i < flag.NArg(); i++ {
			f, err := os.Open(flag.Arg(i))
			if err != nil {
				log.Fatal(err)
			}
			xrep(f, flag.Arg(i))
		}
	}

	err = out.Flush()
	if err != nil {
		logFatal(err)
	}
}

func xrep(f io.Reader, filename string) {
	r := &reader{r: bufio.NewReader(f), i: -1}

	var n, c int
	for {
		n++
		ln := xre.FindReaderIndex(r)
		if ln == nil {
			break
		}

		var section []byte
		if yy {
			section = runes(r.buf[:ln[0]])
		} else {
			section = runes(r.buf[ln[0]:ln[1]])
		}

		if o {
			c++
			p(filename, 0, []byte(string(section) + "\n"))
		} else {
			if xor(re.Match(section), v) {
				c++
				p(filename, n, section)
			}
		}

		r.i = 0
		r.buf = r.buf[ln[1]:]
	}

	if wantc {
		if !h {
			write([]byte(filename + ":"))
		}
		write([]byte(strconv.Itoa(c) + "\n"))
	}

	if c > 0 {
		ret = 0
		if l {
			write([]byte(filename + "\n"))
		}
	}
}

func p(filename string, n int, b[]byte) {
	if wantc || l {
		return
	}

	if !h {
		write([]byte(filename + ":"))
	}

	if wantn {
		write([]byte(strconv.Itoa(n) + ":"))
	}

	write(b)
}

func runes(p []entry) []byte {
	k := make([]byte, 4)
	b := make([]byte, 0, len(p))
	for _, r := range p {
		size := utf8.EncodeRune(k, r.r)
		b = append(b, k[:size]...)
	}
	return b
}

func write(b []byte) {
	if s == true {
		return
	}

	_, err := out.Write(b)
	if err != nil {
		log.Fatal(err)
	}
}

func logFatal(v ...interface{}) {
	if !s {
		log.Print(v...)
	}

	os.Exit(2)
}

func xor(a, b bool) bool {
	if (a && b) || (!a && !b) {
		return false
	}

	return true
}
