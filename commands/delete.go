package commands

import (
	"github.com/cloudfoundry/cli/plugin"
	. "github.com/jberkhahn/v3_beta/models"
	// . "github.com/jberkhahn/v3_beta/util"
	"encoding/json"
	"fmt"
)

func Delete(cliConnection plugin.CliConnection, args []string) {
	appName := args[0]
	fmt.Printf("Deleting app %s...\n", appName)

	output, _ := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps?names=%s", appName))
	apps := V3AppsModel{}
	json.Unmarshal([]byte(output[0]), &apps)

	if len(apps.Apps) == 0 {
		fmt.Printf("App %s not found\n", appName)
    return
  }

	if _, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s", apps.Apps[0].Guid), "-X", "DELETE"); err != nil {
    fmt.Printf("Failed to delete app %s\n", appName)
    return
	}

	fmt.Println("OK")
}
