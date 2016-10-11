package commands

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
	. "github.com/cloudfoundry/v3-cli-plugin/models"
)

func Delete(cliConnection plugin.CliConnection, args []string) {
	appName := args[1]

	urlValues := url.Values{}
	urlValues.Add("names", args[1])

	fmt.Printf("Deleting app %s...\n", appName)

	rawOutput, _ := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps?%s", urlValues.Encode()))
	apps := V3AppsModel{}
	output := strings.Join(rawOutput, "")
	json.Unmarshal([]byte(output), &apps)

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
