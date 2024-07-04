package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	proxmox "github.com/lukask32/fleeting-plugin-proxmox/cmd/fleeting-plugin-proxmox/plugin"
	"gitlab.com/gitlab-org/fleeting/fleeting/plugin"
)

//go:embed licenses.txt
var licenses string

func main() {
	var (
		showVersion  = flag.Bool("version", false, "Show version information and exit")
		showLicenses = flag.Bool("licenses", false, "Show licenses information and exit")
	)

	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	if *showLicenses {
		printLicenses()
		os.Exit(0)
	}

	plugin.Serve(&proxmox.InstanceGroup{})
}

func printVersion() {
	version := plugin.VersionInfo{
		Name: "fleeting-plugin-proxmox",
	}

	info, ok := debug.ReadBuildInfo()
	if ok {
		version.Version = info.Main.Version
	}

	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			version.Revision = setting.Value
		case "vcs.time":
			version.BuiltAt = setting.Value
		}
	}

	fmt.Println(version.Full()) //nolint:forbidigo
}

func printLicenses() {
	fmt.Println(licenses) //nolint:forbidigo
}
