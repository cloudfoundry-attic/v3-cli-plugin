/**
* This is an example plugin where we use both arguments and flags. The plugin
* will echo all arguments passed to it. The flag -uppercase will upcase the
* arguments passed to the command.
**/
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/cloudfoundry/cli/plugin"
)

type V3Plugin struct {
	uppercase *bool
}

type V3AppsModel struct {
	Name       string
	Guid       string
	Error_Code string
}

type V3PackageModel struct {
	Guid       string
	Error_Code string
}

func main() {
	plugin.Start(new(V3Plugin))
}

func (pluginDemo *V3Plugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "v3-push" && len(args) == 3 {
		push(cliConnection, args)
	}
}

func push(cliConnection plugin.CliConnection, args []string) {
	mySpace, _ := cliConnection.GetCurrentSpace()
	//create the app
	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", "/v3/apps", "-X", "POST", "-d", fmt.Sprintf("{\"name\":\"%s\",\"space_guid\":\"%s\"}", args[1], mySpace.Guid))
	freakOut(err)
	app := V3AppsModel{}
	err = json.Unmarshal([]byte(output[0]), &app)
	freakOut(err)
	if app.Error_Code != "" {
		freakOut(errors.New("Error creating v3 app: " + app.Error_Code))
	}

	//create the empty package to upload the app bits to
	output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/packages", app.Guid), "-X", "POST", "-d", "{\"type\":\"bits\"}")
	freakOut(err)
	token, err := cliConnection.AccessToken()
	freakOut(err)
	api, apiErr := cliConnection.ApiEndpoint()
	freakOut(apiErr)
	pack := V3PackageModel{}
	err = json.Unmarshal([]byte(output[0]), &pack)
	if err != nil {
		freakOut(errors.New("Error creating v3 app package: " + app.Error_Code))
	}

	curlOut, upload := exec.Command("curl", fmt.Sprintf("%s/v3/packages/%s/upload", api, pack.Guid), "-F", fmt.Sprintf("bits=@\"%s\"", args[2]), "-H", fmt.Sprintf("\"Authorization: %s", token)).Output()
	fmt.Println("woo: ", curlOut, api, pack.Guid, args[2])
	freakOut(upload)

}

func freakOut(err error) {
	if err != nil {
		fmt.Println("Error Will Robinson!: ", err.Error())
		os.Exit(1)
	}

}

func (pluginDemo *V3Plugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "V3_API",
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
		},
	}
}
