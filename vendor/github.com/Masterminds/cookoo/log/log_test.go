package log

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/Masterminds/cookoo"
)

func Example() {
	_, _, c := cookoo.Cookoo()

	// Set logging to go to Stdout.
	c.AddLogger("stdout", os.Stdout)

	// Set the log level to any of the Log* constants.
	Level = LogInfo

	// Functions are named as they are in log/syslog.
	Err(c, "Failed to do something.")

	// There are also formatting versions of every log function.
	Infof(c, "Informational message with %s.", "formatting")

	// Shorthand for if Level >= LogDebug
	if Debugging() {
		Debug(c, "This is a debug message.")
	}

	// You can test for any level.
	if Level >= LogWarning {
		Warn(c, "And this is a warning.")
	}

	Stack(c, "Stack trace from here.")
}

func TestLog(t *testing.T) {
	_, _, c := cookoo.Cookoo()
	var b bytes.Buffer
	c.AddLogger("stdout", &b)
	Errf(c, "Failed %s", "now")

	if !strings.Contains(b.String(), "Failed now") {
		t.Fatalf("Expected '%s' to contain 'Failed now'", b.String())
	}

	if Level != LogDebug {
		t.Errorf("Expected log level %d, got %d", LogDebug, Level)
	}

	if Label[Level] != LabelDebug {
		t.Errorf("Expected log label '%s', got '%s'", LabelDebug, Label[Level])
	}

	Level = LogErr
	t.Logf("Set level to %d", LogErr)

	b.Reset()
	if LogDebug <= Level {
		Debug(c, "foo")
	}
	if b.Len() > 0 {
		t.Errorf("Expected empty buffer. Got %s", b.String)
	}

	Alert(c, "test")
	if !strings.Contains(b.String(), "test") {
		t.Errorf("Expected alert to get logged on log level %s", Label[Level])
	}

	Stack(c, "Stack trace from here.")
	t.Log(b.String())
}
