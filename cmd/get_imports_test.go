package cmd

import "github.com/Masterminds/cookoo"

func SilenceLogs(c cookoo.Context) {
	p := cookoo.NewParamsWithValues(map[string]interface{}{"quiet": true})
	BeQuiet(c, p)
}
