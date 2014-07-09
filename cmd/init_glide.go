package cmd

import (
	"github.com/Masterminds/cookoo"
	"fmt"
)

func InitGlide(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	fmt.Printf("[INFO] Initialized. You can now edit 'glide.yaml'\n")
	return true, nil
}
