package commands

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/plugin"
	"github.com/cloudfoundry/v3-cli-plugin/models"
	"github.com/cloudfoundry/v3-cli-plugin/util"
)

type runningTask struct {
	guid    string
	command string
	state   string
	time    time.Time
}

func CancelTask(cliConnection plugin.CliConnection, args []string) {
	appName := args[1]
	taskName := args[2]

	rawOutput, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps?names=%s", appName))
	util.FreakOut(err)
	output := strings.Join(rawOutput, "")

	apps := models.V3AppsModel{}
	json.Unmarshal([]byte(output), &apps)

	if len(apps.Apps) == 0 {
		fmt.Printf("App %s not found\n", appName)
		return
	}

	appGuid := apps.Apps[0].Guid

	rawTasksJson, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/tasks", appGuid))
	util.FreakOut(err)
	tasksJson := strings.Join(rawTasksJson, "")

	tasks := models.V3TasksModel{}
	err = json.Unmarshal([]byte(tasksJson), &tasks)
	util.FreakOut(err)

	var runningTasks []runningTask
	for _, task := range tasks.Tasks {
		if taskName == task.Name && task.State == "RUNNING" {
			runningTasks = append(runningTasks, runningTask{task.Guid, task.Command, task.State, task.UpdatedAt})
		}
	}

	if len(runningTasks) == 0 {
		fmt.Println("No running task found. Task name:", taskName)
		return
	} else if len(runningTasks) == 1 {
		rawOutput, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/tasks/%s/cancel", runningTasks[0].guid), "-X", "PUT", "-d", "{}")
		util.FreakOut(err)
		output := strings.Join(rawOutput, "")
		fmt.Println(output)
		return
	} else {
		fmt.Printf("Please select which task to cancel: \n\n")
		tasksTable := util.NewTable([]string{"#", "Task Name", "Command", "State", "Time"})
		for i, task := range runningTasks {
			tasksTable.Add(
				strconv.Itoa(i+1),
				taskName,
				task.command,
				task.state,
				fmt.Sprintf("%s", task.time.Format("Jan 2, 15:04:05 MST")),
			)
		}
		tasksTable.Print()

		var i int64 = -1
		var str string

		for i <= 0 || i > int64(len(runningTasks)) {
			fmt.Printf("\nSelect from above > ")
			fmt.Scanf("%s", &str)
			i, _ = strconv.ParseInt(str, 10, 32)
		}

		rawOutput, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/tasks/%s/cancel", runningTasks[i-1].guid), "-X", "PUT", "-d", "{}")
		util.FreakOut(err)
		output := strings.Join(rawOutput, "")
		fmt.Println(output)
	}
}
