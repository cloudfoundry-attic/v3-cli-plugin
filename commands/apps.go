package commands

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cloudfoundry/cli/plugin"
	. "github.com/cloudfoundry/v3-cli-plugin/models"
	"github.com/cloudfoundry/v3-cli-plugin/util"
)

func Apps(cliConnection plugin.CliConnection, args []string) {
	mySpace, err := cliConnection.GetCurrentSpace()
	util.FreakOut(err)
	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("v3/apps?space_guids=%s", mySpace.Guid), "-X", "GET")
	util.FreakOut(err)
	apps := V3AppsModel{}
	err = json.Unmarshal([]byte(output[0]), &apps)
	util.FreakOut(err)

	if len(apps.Apps) > 0 {
		appsTable := util.NewTable([]string{("name"), ("total_desired_instances")})
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
