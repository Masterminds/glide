package safely

import (
	"testing"
	"bytes"
	"log"
	"time"
)

func TestGo( t *testing.T) {
	f := func() {
		panic("OUCH!")
	}
	Go(f)
}

func TestGoLog(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	GoLog(logger, func() {
		panic("OUCH")
	})
	time.Sleep(time.Second)
	if !bytes.Contains(buf.Bytes(), []byte("OUCH")) {
		t.Errorf("Expected to find 'OUCH' in %s.", buf.String())
	}
}
