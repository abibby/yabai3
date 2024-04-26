package bar

import "io"

type PrefixWriter struct {
	w         io.Writer
	prefix    string
	addPrefix bool
}

func NewPrefixWriter(w io.Writer, prefix string) *PrefixWriter {
	return &PrefixWriter{
		w:         w,
		prefix:    prefix,
		addPrefix: true,
	}
}

func (w *PrefixWriter) Write(b []byte) (n int, err error) {
	buff := make([]byte, 0, len(b)+len(w.prefix))
	for _, c := range b {
		if w.addPrefix {
			w.addPrefix = false
			buff = append(buff, []byte(w.prefix)...)
		}
		buff = append(buff, c)
		if c == '\n' {
			w.addPrefix = true
		}
	}
	return w.w.Write(buff)
}

func writeAll(w io.Writer, bytes ...[]byte) error {
	for _, b := range bytes {
		_, err := w.Write(b)
		if err != nil {
			return err
		}
	}
	return nil
}
