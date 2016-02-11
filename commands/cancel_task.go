package commands

import (
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry/cli/plugin"
	"github.com/jberkhahn/v3_beta/models"
	"github.com/jberkhahn/v3_beta/util"
)

func CancelTask(cliConnection plugin.CliConnection, args []string) {
	appName := args[1]
	taskName := args[2]

	output, _ := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps?names=%s", appName))
	apps := models.V3AppsModel{}
	json.Unmarshal([]byte(output[0]), &apps)

	if len(apps.Apps) == 0 {
		fmt.Printf("App %s not found\n", appName)
		return
	}

	appGuid := apps.Apps[0].Guid

	tasksJson, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/tasks", appGuid))
	util.FreakOut(err)

	tasks := models.V3TasksModel{}
	err = json.Unmarshal([]byte(tasksJson[0]), &tasks)
	util.FreakOut(err)

	for _, task := range tasks.Tasks {
		if taskName == task.Name {
			output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/tasks/%s/cancel", task.Guid), "-X", "PUT", "-d", "{}")
			util.FreakOut(err)

			fmt.Println("output:", output)
			return
		}
	}

	fmt.Println("No task found. Task name:", taskName)
}
