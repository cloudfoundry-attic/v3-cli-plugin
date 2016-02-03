package commands

import (
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry/cli/plugin"
	. "github.com/jberkhahn/v3_beta/models"
)

func RunTask(cliConnection plugin.CliConnection, args []string) {
	appName := args[1]
	taskName := args[2]
	taskCommand := args[3]

	fmt.Printf("Running task %s on app %s...\n", taskName, appName)

	output, _ := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps?names=%s", appName))
	apps := V3AppsModel{}
	json.Unmarshal([]byte(output[0]), &apps)

	if len(apps.Apps) == 0 {
		fmt.Printf("App %s not found\n", appName)
		return
	}

	appGuid := apps.Apps[0].Guid

	body := fmt.Sprintf(`{
		"name": "%s", 
		"command": "%s"
	}`, taskName, taskCommand)

	if _, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/tasks", appGuid), "-X", "POST", "-d", body); err != nil {
		fmt.Printf("Failed to run task %s\n", taskName)
		return
	}

	fmt.Println("OK")
}
