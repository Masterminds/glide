package action

import (
	"fmt"
)

func Release() {
	conf := EnsureConfig()
	fmt.Println(conf.Version)
}
