package cmd

import (
	"fmt"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/glide/cfg"
)

// PrintName prints the name of the project.
//
// This comes from Config.Name.
func PrintName(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	conf := p.Get("conf", nil).(*cfg.Config)
	fmt.Println(conf.Name)
	return nil, nil
}
