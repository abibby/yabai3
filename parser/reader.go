package parser

import "os"

const EOF = byte(0xff)

type Reader struct {
	file       string
	source     []byte
	cursor     int
	whitespace []byte
	line       int
	column     int
}

type ReaderTx struct {
	reader    *Reader
	cursor    int
	line      int
	column    int
	committed bool
}

func (r *Reader) BeginTx() *ReaderTx {
	return &ReaderTx{
		reader:    r,
		cursor:    r.cursor,
		line:      r.line,
		column:    r.column,
		committed: false,
	}
}

func (tx *ReaderTx) Commit() *Section {
	tx.committed = true
	return NewSection(tx.cursor, tx.reader.cursor-tx.cursor)
}

func (tx *ReaderTx) Rollback() {
	if tx.committed {
		return
	}
	tx.reader.cursor = tx.cursor
	tx.reader.line = tx.line
	tx.reader.column = tx.column
}

func NewReaderFromFile(file string) (*Reader, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return NewReader(file, b), nil
}
func NewReader(file string, b []byte) *Reader {
	return &Reader{
		file:       file,
		source:     b,
		cursor:     0,
		line:       1,
		column:     1,
		whitespace: []byte(" \n\t"),
	}
}

func (r *Reader) isWhitespace(b byte) bool {
	for _, w := range r.whitespace {
		if w == b {
			return true
		}
	}
	return false
}

func (r *Reader) Source() []byte {
	return r.source
}

func (r *Reader) Peak() byte {
	if r.cursor >= len(r.source) {
		return EOF
	}
	return r.source[r.cursor]
}

func (r *Reader) ReadByte() byte {
	b := r.Peak()
	r.Advance(1)
	return b
}
func (r *Reader) SkipWhitespace() {
	b := r.Peak()
	for r.isWhitespace(b) {
		r.Advance(1)
		b = r.Peak()
	}
}
func (r *Reader) PeakWord() []byte {
	return r.PeakUntil([]byte(" \t\n"))
}
func (r *Reader) ReadWord() []byte {
	w := r.PeakUntil([]byte(" \t\n"))
	r.Advance(len(w))
	return w
}

func (r *Reader) PeakN(length int) []byte {
	return r.source[r.cursor : r.cursor+length]
}
func (r *Reader) ReadN(length int) []byte {
	b := r.PeakN(length)
	r.Advance(len(b))
	return b
}

func (r *Reader) PeakLine() []byte {
	return r.PeakUntil([]byte("\n"))
}
func (r *Reader) ReadUntil(set []byte) []byte {
	b := r.PeakUntil(set)
	r.Advance(len(b))
	return b
}
func (r *Reader) PeakUntil(set []byte) []byte {
	for i := r.cursor; i < len(r.source); i++ {
		for _, b := range set {
			if r.source[i] == b {
				return r.source[r.cursor:i]
			}
		}
	}
	return r.source[r.cursor:]
}
func (r *Reader) Advance(count int) {
	b := r.PeakN(count)
	for _, c := range b {
		if c == '\n' {
			r.line++
			r.column = 1
		} else {
			r.column++
		}

	}
	r.cursor = r.cursor + count
}

func (r *Reader) File() string {
	return r.file
}
func (r *Reader) Line() int {
	return r.line
}
func (r *Reader) Column() int {
	return r.column
}
