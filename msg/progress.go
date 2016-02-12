package msg

import (
	"fmt"
	"time"
)

var Cylon = []string{
	" 〖\033[0;31m        ◼︎\033[m〗",
	" 〖\033[0;31m       ◼︎ \033[m〗",
	" 〖\033[0;31m      ◼︎▫︎ \033[m〗",
	" 〖\033[0;31m     ◼︎ ▫︎▫︎\033[m〗",
	" 〖\033[0;31m    ◼︎ ▫︎▫︎ \033[m〗",
	" 〖\033[0;31m   ◼︎ ▫︎▫︎  \033[m〗",
	" 〖\033[0;31m  ◼︎ ▫︎▫︎   \033[m〗",
	" 〖\033[0;31m ◼︎ ▫︎▫︎    \033[m〗",
	" 〖\033[0;31m◼︎ ▫︎▫︎     \033[m〗",
	" 〖\033[0;31m◼︎        \033[m〗",
	" 〖\033[0;31m▫︎◼︎       \033[m〗",
	" 〖\033[0;31m▫︎▫︎◼︎      \033[m〗",
	" 〖\033[0;31m ▫︎▫︎◼︎     \033[m〗",
	" 〖\033[0;31m  ▫︎▫︎◼︎    \033[m〗",
	" 〖\033[0;31m   ▫︎▫︎◼︎   \033[m〗",
	" 〖\033[0;31m    ▫︎▫︎◼︎  \033[m〗",
	" 〖\033[0;31m      ▫︎▫︎◼︎\033[m〗",
}

const CylonInterval = 100 * time.Millisecond

func StartProgress(msg string) {
	Default.InProgress = true
	Default.meter.Start(msg)
}

func Progress(msg string, v ...interface{}) {
	if Default.InProgress {
		Default.meter.Message(fmt.Sprintf(msg, v...))
	}
}

func StopProgress(msg string) {
	Default.meter.Done(msg)
	Default.InProgress = false
}

// nilMeter is an implementation of ProgressMeter that prints messages to Info.
type nilMeter int

func (n nilMeter) Start(s string) {
	Info(s)
}
func (n nilMeter) Message(s string) {
	Info(s)
}
func (n nilMeter) Done(s string) {
	Info(s)
}
