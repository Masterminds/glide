package convert
/*
Conversion commands.

Convert one type to another.

This package provides conversions that can be used in chains of commands.
*/

import (
	"github.com/Masterminds/cookoo"
	"strconv"
)

// Convert a string to an integer.
//
// Params:
// 	- str (string): A string that contains a number.
//
// Returns:
// 	- An integer.
func Atoi(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	src := p.Get("str", "0").(string)
	return strconv.Atoi(src)
}
