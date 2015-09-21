package commands

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
	. "github.com/jberkhahn/v3_beta/models"
	. "github.com/jberkhahn/v3_beta/util"
)

func Processes(cliConnection plugin.CliConnection, args []string) {
	mySpace, err := cliConnection.GetCurrentSpace()
	FreakOut(err)

	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", "v3/processes?per_page=5000", "-X", "GET")
	FreakOut(err)
	processes := V3ProcessesModel{}
	err = json.Unmarshal([]byte(output[0]), &processes)
	FreakOut(err)

	if len(processes.Processes) > 0 {
		processesTable := NewTable([]string{("app"), ("type"), ("memory in MB"), ("disk in MB")})
		for _, v := range processes.Processes {
			if strings.Contains(v.Links.Space.Href, mySpace.Guid) {
				appName := "N/A"
				if v.Links.App.Href != "/v3/apps/" {
					appName = strings.Split(v.Links.App.Href, "/v3/apps/")[1]
				}
				processesTable.Add(
					appName,
					v.Type,
					strconv.Itoa(v.Memory)+"MB",
					strconv.Itoa(v.Disk)+"MB",
				)
			}
		}
		fmt.Println("print table?")
		processesTable.Print()
	} else {
		fmt.Println("No v3 processes found.")
	}
}
