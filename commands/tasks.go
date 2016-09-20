package commands

import (
	"encoding/json"
	"fmt"
	"strings"
	"strconv"

	"github.com/cloudfoundry/cli/plugin"
	. "github.com/cloudfoundry/v3-cli-plugin/models"
	. "github.com/cloudfoundry/v3-cli-plugin/util"
)

func Tasks(cliConnection plugin.CliConnection, args []string) {
	appName := args[1]
	fmt.Printf("Listing tasks for app %s...\n", appName)

	rawOutput, _ := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps?names=%s", appName))
	output := strings.Join(rawOutput, "")
	apps := V3AppsModel{}
	json.Unmarshal([]byte(output), &apps)

	if len(apps.Apps) == 0 {
		fmt.Printf("App %s not found\n", appName)
		return
	}

	appGuid := apps.Apps[0].Guid

	rawOutput, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/tasks", appGuid), "-X", "GET")
	FreakOut(err)
	output = strings.Join(rawOutput, "")
	tasks := V3TasksModel{}
	err = json.Unmarshal([]byte(output), &tasks)
	FreakOut(err)

	if len(tasks.Tasks) > 0 {
		tasksTable := NewTable([]string{("id"), ("name"), ("command"), ("state")})
		for _, v := range tasks.Tasks {
			tasksTable.Add(
				strconv.Itoa(v.Id),
				v.Name,
				v.Command,
				v.State,
			)
		}
		tasksTable.Print()
	} else {
		fmt.Println("No v3 tasks found.")
	}
}
