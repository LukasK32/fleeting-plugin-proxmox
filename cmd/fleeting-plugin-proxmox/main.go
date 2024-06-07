package main

import (
	proxmox "github.com/lukask32/fleeting-plugin-proxmox/cmd/fleeting-plugin-proxmox/plugin"
	"gitlab.com/gitlab-org/fleeting/fleeting/plugin"
)

func main() {
	plugin.Main(
		&proxmox.InstanceGroup{},
		plugin.VersionInfo{
			Name: "fleeting-plugin-proxmox",
		},
	)
}
