package action

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/Masterminds/glide/msg"
)

func TestList(t *testing.T) {
	msg.Default.PanicOnDie = true
	old := msg.Default.Stdout
	defer func() {
		msg.Default.Stdout = old
	}()

	var buf bytes.Buffer
	msg.Default.Stdout = &buf
	List("../", false, "text")
	if buf.Len() < 5 {
		t.Error("Expected some data to be found.")
	}

	var buf2 bytes.Buffer
	msg.Default.Stdout = &buf2
	List("../", false, "json")
	j := buf2.Bytes()
	var o PackageList
	err := json.Unmarshal(j, &o)
	if err != nil {
		t.Errorf("Error unmarshaling json list: %s", err)
	}
	if len(o.Installed) == 0 {
		t.Error("No packages found on json list")
	}

	var buf3 bytes.Buffer
	msg.Default.Stdout = &buf3
	List("../", false, "json-pretty")
	j = buf3.Bytes()
	var o2 PackageList
	err = json.Unmarshal(j, &o2)
	if err != nil {
		t.Errorf("Error unmarshaling json-pretty list: %s", err)
	}
	if len(o2.Installed) == 0 {
		t.Error("No packages found on json-pretty list")
	}
}
