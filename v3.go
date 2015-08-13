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
	"strings"
	"time"

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

type V3DropletModel struct {
	Guid string
}

type MetadataModel struct {
	Guid string `json:"guid"`
}

type EntityModel struct {
	Name string `json:"name"`
}
type RouteEntityModel struct {
	Host string `json:"host"`
}

type DomainsModel struct {
	NextUrl   string        `json:"next_url,omitempty"`
	Resources []DomainModel `json:"resources"`
}
type DomainModel struct {
	Metadata MetadataModel `json:"metadata"`
	Entity   EntityModel   `json:"entity"`
}
type RouteModel struct {
	Metadata MetadataModel    `json:"metadata"`
	Entity   RouteEntityModel `json:"entity"`
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

	apiString := fmt.Sprintf("%s", api)
	if strings.Index(apiString, "s") == 4 {
		apiString = apiString[:4] + apiString[5:]
	}
	_, upload := exec.Command("curl", fmt.Sprintf("%s/v3/packages/%s/upload", apiString, pack.Guid), "-F", fmt.Sprintf("bits=@%s", args[2]), "-H", fmt.Sprintf("Authorization: %s", token)).Output()
	freakOut(upload)

	//need to sleep or the CCDB won't be updated before with the bits before we try to make the droplet
	//make a polling loop here maybe?
	time.Sleep(5 * time.Second)
	output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/packages/%s/droplets", pack.Guid), "-X", "POST", "-d", "{}")
	freakOut(err)
	droplet := V3DropletModel{}
	err = json.Unmarshal([]byte(output[0]), &droplet)
	if err != nil {
		freakOut(errors.New("error marshaling the v3 droplet: " + err.Error()))
	}

	//cf curl /v3/apps/[your-app-guid]/current_droplet -X PUT -d '{"droplet_guid": "[your-droplet-guid]"}'
	//make a polling loop here maybe?
	time.Sleep(10 * time.Second)
	output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/current_droplet", app.Guid), "-X", "PUT", "-d", fmt.Sprintf("{\"droplet_guid\":\"%s\"}", droplet.Guid))
	freakOut(err)

	//CF_TRACE=true cf create-route [space-name] [domain-name] -n [host-name]
	space, _ := cliConnection.GetCurrentSpace()
	nextUrl := "/v2/shared_domains"
	allDomains := DomainsModel{}
	for nextUrl != "" {
		output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", nextUrl)
		freakOut(err)
		tmp := DomainsModel{}
		err = json.Unmarshal([]byte(output[0]), &tmp)
		freakOut(err)
		allDomains.Resources = append(allDomains.Resources, tmp.Resources...)

		if tmp.NextUrl != "" {
			nextUrl = tmp.NextUrl
		} else {
			nextUrl = ""
		}
	}
	domainGuid := ""
	for _, v := range allDomains.Resources {
		if v.Entity.Name == apiString[11:] {
			domainGuid = v.Metadata.Guid
		}
	}
	output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", "v2/routes", "-X", "POST", "-d", fmt.Sprintf(`{"host":"%s","domain_guid":"%s","space_guid":"%s"}`, args[1], domainGuid, space.Guid))
	freakOut(err)
	route := RouteModel{}
	err = json.Unmarshal([]byte(output[0]), &route)
	if err != nil {
		freakOut(errors.New("error unmarshaling the route: " + err.Error()))
	}

	//map the route to the app
	time.Sleep(1 * time.Second)
	output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/routes", app.Guid), "-X", "PUT", "-d", fmt.Sprintf("{\"route_guid\": \"%s\"}", route.Metadata.Guid))
	freakOut(err)

	//start the app
	output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/start", app.Guid), "-X", "PUT")
	freakOut(err)

	fmt.Println("Done pushing! Checkout your processes using 'cf apps'")
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
