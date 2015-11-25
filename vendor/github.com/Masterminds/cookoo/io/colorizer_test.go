package io

import (
	"bytes"
	"testing"
)

func TestColorizer(t *testing.T) {

	buffer := new(bytes.Buffer)
	colorizer := NewColorizer(buffer)
	msg := []byte("error test")

	colorizer.Write(msg)
	line := buffer.String()
	if line != "\033[0;31merror test\033[m" {
		t.Errorf("! Error message was not colorized.")
	}

	buffer = new(bytes.Buffer)
	colorizer = NewColorizer(buffer)
	msg = []byte("warn test")
	colorizer.Write(msg)
	line = buffer.String()
	if line != "\033[0;33mwarn test\033[m" {
		t.Errorf("! Warn message was not colorized.")
	}

	buffer = new(bytes.Buffer)
	colorizer = NewColorizer(buffer)
	msg = []byte("info test")
	colorizer.Write(msg)
	line = buffer.String()
	if line != "\033[0;36minfo test\033[m" {
		t.Errorf("! Info message was not colorized.")
	}

}
