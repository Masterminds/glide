package io

import (
	"io"
	"os"
	"fmt"
)

// MultiWriter enables you to have a writer that passes on the writing to one
// of more Writers where the write is duplicated to each Writer. MultiWriter
// is similar to the multiWriter that is part of Go. The difference is
// this MultiWriter allows you to manager the Writers attached to it via CRUD
// operations. To do this you will need to mock the type. For example,
// mw := NewMultiWriter()
// mw.(*MultiWriter).AddWriter("foo", foo)
type MultiWriter struct {
	writers map[string]io.Writer
}

// Write sends the bytes to each of the attached writers to be written.
func (t *MultiWriter) Write(p []byte) (n int, err error) {
	for name, w := range t.writers {
		n, err = w.Write(p)
		if err != nil {
			// One broken logger should not stop the others.
			fmt.Fprintf(os.Stderr, "Error logging to '%s': %s", name, err)
			continue
		}
		if n < len(p) {
			// One broken logger should not stop the others.
			err = io.ErrShortWrite
			fmt.Fprintf(os.Stderr, "Short write logging to '%s': Expected to write %d (%V), wrote %d", name, len(p), w, n)
			continue
		}
	}
	return len(p), err
}

// Init initializes the MultiWriter.
func (t *MultiWriter) Init() *MultiWriter {
	t.writers = make(map[string]io.Writer)
	return t
}

// Writer retrieves a given io.Writer given its name.
func (t *MultiWriter) Writer(name string) (io.Writer, bool) {
	value, found := t.writers[name]
	return value, found
}

// Writers retrieves a map of all io.Writers keyed by name.
func (t *MultiWriter) Writers() map[string]io.Writer {
	return t.writers
}

// AddWriter adds an io.Writer with an associated name.
func (t *MultiWriter) AddWriter(name string, writer io.Writer) {
	t.writers[name] = writer
}

// RemoveWriter removes an io.Writer given a name.
func (t *MultiWriter) RemoveWriter(name string) {
	delete(t.writers, name)
}

// NewMultiWriter returns an initialized MultiWriter.
func NewMultiWriter() io.Writer {
	w := new(MultiWriter).Init()
	return w
}
