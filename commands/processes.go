package commands

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
	. "github.com/cloudfoundry/v3-cli-plugin/models"
	. "github.com/cloudfoundry/v3-cli-plugin/util"
)

func Processes(cliConnection plugin.CliConnection, args []string) {
	mySpace, err := cliConnection.GetCurrentSpace()
	FreakOut(err)

	rawOutput, err := cliConnection.CliCommandWithoutTerminalOutput("curl", "v3/processes?per_page=5000", "-X", "GET")
	FreakOut(err)
	output := strings.Join(rawOutput, "")
	processes := V3ProcessesModel{}
	err = json.Unmarshal([]byte(output), &processes)
	FreakOut(err)

	rawOutput, err = cliConnection.CliCommandWithoutTerminalOutput("curl", "v3/apps?per_page=5000", "-X", "GET")
	FreakOut(err)
	output = strings.Join(rawOutput, "")
	apps := V3AppsModel{}
	err = json.Unmarshal([]byte(output), &apps)
	FreakOut(err)
	appsMap := make(map[string]V3AppModel)
	for _, app := range apps.Apps {
		appsMap[app.Guid] = app
	}

	if len(processes.Processes) > 0 {
		processesTable := NewTable([]string{("app"), ("type"), ("instances"), ("memory in MB"), ("disk in MB")})
		for _, v := range processes.Processes {
			if strings.Contains(v.Links.Space.Href, mySpace.Guid) {
				appName := "N/A"
				if v.Links.App.Href != "/v3/apps/" {
					appGuid := strings.Split(v.Links.App.Href, "/v3/apps/")[1]
					appName = appsMap[appGuid].Name
				}
				processesTable.Add(
					appName,
					v.Type,
					strconv.Itoa(v.Instances),
					strconv.Itoa(v.Memory)+"MB",
					strconv.Itoa(v.Disk)+"MB",
				)
			}
		}
		processesTable.Print()
	} else {
		fmt.Println("No v3 processes found.")
	}
}
