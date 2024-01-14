package main

import (
	"flag"
	"fmt"
	"os"
)

// Injected during CI build
var (
	VERSION = "dev"
)

var (
	showVersion = flag.Bool("version", false, "Show version and exit")
)

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	fmt.Println("Hello World!")
}
