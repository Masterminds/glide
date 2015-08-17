package cmd

import (
	"fmt"

	"github.com/Masterminds/cookoo"
)

func PrintName(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cfg := p.Get("conf", nil).(*Config)
	fmt.Println(cfg.Name)
	return nil, nil
}
