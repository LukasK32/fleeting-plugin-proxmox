package main

import (
	"flag"
	"fmt"
	"os"

	proxmox "gitlab.com/LukasK32/fleeting-plugin-proxmox/cmd/fleeting-plugin-proxmox/plugin"
	"gitlab.com/gitlab-org/fleeting/fleeting/plugin"
)

// Injected during CI build.
var (
	//nolint:gochecknoglobals
	VERSION = "dev"
)

func main() {
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *showVersion {
		//nolint:forbidigo
		fmt.Println(VERSION)
		os.Exit(0)
	}

	plugin.Serve(&proxmox.InstanceGroup{})
}
