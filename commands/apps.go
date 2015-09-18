package commands

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cloudfoundry/cli/plugin"
	. "github.com/jberkhahn/v3_beta/models"
	. "github.com/jberkhahn/v3_beta/util"
)

func Apps(cliConnection plugin.CliConnection, args []string) {
	//mySpace, err := cliConnection.GetCurrentSpace()
	//FreakOut(err)
	//output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("v3/apps&space_guid=%s", mySpace.Guid), "-X", "GET")

	//currently global - get all v3 apps regardless of space
	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", "v3/apps", "-X", "GET")
	FreakOut(err)
	apps := V3AppsModel{}
	err = json.Unmarshal([]byte(output[0]), &apps)
	FreakOut(err)

	if len(apps.Apps) > 0 {
		appsTable := NewTable([]string{("name"), ("total_desired_instances")})
		for _, v := range apps.Apps {
			appsTable.Add(
				v.Name,
				strconv.Itoa(v.Instances),
			)
		}
		appsTable.Print()
	} else {
		fmt.Println("No v3 apps found.")
	}
}
