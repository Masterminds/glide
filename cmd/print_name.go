package cmd

import (
	"fmt"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/glide/yaml"
)

// PrintName prints the name of the project.
//
// This comes from Config.Name.
func PrintName(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cfg := p.Get("conf", nil).(*yaml.Config)
	fmt.Println(cfg.Name)
	return nil, nil
}
