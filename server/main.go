package main

import (
	"fmt"
	"github.com/c-mueller/serverless-doh/config"
)

func main() {
	fmt.Println(len(config.Blocklists))
}
