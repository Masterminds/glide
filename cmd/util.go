package cmd

import (
	"github.com/Masterminds/cookoo"
	"fmt"
)

var Quiet bool = false

func BeQuiet(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	qstr := p.Get("quiet", "false").(string)
	Quiet = qstr == "true"
	return Quiet, nil
}

func Info(msg string, args ...interface{}) {
	if Quiet { return }
	fmt.Print("[INFO] ")
	Msg(msg, args...)
}
func Debug(msg string, args ...interface{}) {
	if Quiet { return }
	fmt.Print("[DEBUG] ")
	Msg(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	fmt.Print("[WARN] ")
	Msg(msg, args...)
}

func Error(msg string, args ...interface{}) {
	fmt.Print("[ERROR] ")
	Msg(msg, args...)
}

func Msg(msg string, args...interface{}) {
	if len(args) == 0 {
		fmt.Print(msg)
		return
	}
	fmt.Printf(msg, args...)
}
