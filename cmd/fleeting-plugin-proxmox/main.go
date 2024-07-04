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
		showHelp     = flag.Bool("help", false, "Show help and exit")
		showVersion  = flag.Bool("version", false, "Show version information and exit")
		showLicenses = flag.Bool("licenses", false, "Show licenses information and exit")
	)

	flag.Parse()

	if *showHelp {
		printHelp()
		os.Exit(0)
	}

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

func printHelp() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
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
