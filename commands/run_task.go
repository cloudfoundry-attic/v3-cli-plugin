package commands

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/plugin"
	"github.com/cloudfoundry/v3-cli-plugin/models"
	"github.com/cloudfoundry/v3-cli-plugin/util"
)

func RunTask(cliConnection plugin.CliConnection, args []string) {
	appName := args[1]
	taskName := args[2]
	taskCommand := args[3]

	fmt.Println("OK\n")
	fmt.Printf("Running task %s on app %s...\n\n", taskName, appName)

	go Logs(cliConnection, args)
	time.Sleep(2 * time.Second) // b/c sharing the cliConnection makes things break

	rawOutput, _ := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps?names=%s", appName))
	output := strings.Join(rawOutput, "")
	apps := models.V3AppsModel{}
	json.Unmarshal([]byte(output), &apps)

	if len(apps.Apps) == 0 {
		fmt.Printf("App %s not found\n", appName)
		return
	}
	appGuid := apps.Apps[0].Guid

	body := fmt.Sprintf(`{
		"name": "%s",
		"command": "%s"
	}`, taskName, taskCommand)

	rawOutput, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/tasks", appGuid), "-X", "POST", "-d", body)
	if err != nil {
		fmt.Printf("Failed to run task %s\n", taskName)
		return
	}
	output = strings.Join(rawOutput, "")

	task := models.V3TaskModel{}
	err = json.Unmarshal([]byte(output), &task)
	util.FreakOut(err)
	if task.Guid == "" {
		fmt.Printf("Failed to run task %s:\n%s\n", taskName, output)
		return
	}

	util.Poll(cliConnection, fmt.Sprintf("/v3/tasks/%s", task.Guid), "SUCCEEDED", 1*time.Minute, "Task failed to run")

	fmt.Printf("Task %s successfully completed.\n", task.Guid)
}
