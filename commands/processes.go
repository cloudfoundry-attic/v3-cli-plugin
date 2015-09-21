package commands

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cloudfoundry/cli/plugin"
	. "github.com/jberkhahn/v3_beta/models"
	. "github.com/jberkhahn/v3_beta/util"
)

func Processes(cliConnection plugin.CliConnection, args []string) {
	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", "v3/processes?per_page=1000", "-X", "GET")
	FreakOut(err)
	processes := V3ProcessesModel{}
	err = json.Unmarshal([]byte(output[0]), &processes)
	FreakOut(err)

	if len(processes.Processes) > 0 {
		processesTable := NewTable([]string{("type"), ("memory in MB"), ("disk in MB")})
		for _, v := range processes.Processes {
			processesTable.Add(
				v.Type,
				strconv.Itoa(v.Memory)+"MB",
				strconv.Itoa(v.Disk)+"MB",
			)
		}
		processesTable.Print()
	} else {
		fmt.Println("No v3 processes found.")
	}
}
