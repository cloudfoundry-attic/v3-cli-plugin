package main

import (
	"github.com/cloudfoundry/cli/plugin"
	"github.com/jberkhahn/v3_beta/commands"
)

type V3Plugin struct {
}

func main() {
	plugin.Start(new(V3Plugin))
}

func (pluginDemo *V3Plugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "v3-push" {
		if len(args) == 2 || len(args) == 3 {
			commands.Push(cliConnection, args)
		} else {
			//print push help
		}
	} else if args[0] == "v3-apps" {
		if len(args) == 1 {
			commands.Apps(cliConnection, args)
		} else {
			//print apps help
		}
	} else if args[0] == "v3-processes" {
		if len(args) == 1 {
			commands.Processes(cliConnection, args)
		} else {
			//print processes help
		}
	}
}

func (pluginDemo *V3Plugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "v3_beta",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 1,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "v3-push",
				Alias:    "",
				HelpText: "pushes a zipped app as a v3 process",
				UsageDetails: plugin.Usage{
					Usage:   "v3-push APPNAME PATH/TO/ZIPPED/APP",
					Options: map[string]string{},
				},
			},
			{
				Name:     "v3-apps",
				Alias:    "v3-a",
				HelpText: "displays all v3 apps",
				UsageDetails: plugin.Usage{
					Usage:   "v3-apps",
					Options: map[string]string{},
				},
			},
			{
				Name:     "v3-processes",
				Alias:    "v3-p",
				HelpText: "displays all v3 processes",
				UsageDetails: plugin.Usage{
					Usage:   "v3-processes",
					Options: map[string]string{},
				},
			},
		},
	}
}
