package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
	. "github.com/cloudfoundry/v3-cli-plugin/models"
	. "github.com/cloudfoundry/v3-cli-plugin/util"
	"github.com/simonleung8/flags"
)

func BindService(cliConnection plugin.CliConnection, args []string) {
	customParameters := "{}"
	fc := flags.New()
	fc.NewStringFlag("parameters", "c", "Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering.")
	fc.Parse(args...)
	if fc.IsSet("c") {
		customParameters = fc.String("c")
	}

	appName := fc.Args()[1]
	serviceInstanceName := fc.Args()[2]

	rawOutput, _ := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps?names=%s", appName))
	apps := V3AppsModel{}
	output := strings.Join(rawOutput, "")
	json.Unmarshal([]byte(output), &apps)

	if len(apps.Apps) == 0 {
		fmt.Printf("App %s not found\n", appName)
		return
	}

	appGuid := apps.Apps[0].Guid

	serviceInstance, err := cliConnection.GetService(serviceInstanceName)
	FreakOut(err)
	serviceInstanceGuid := serviceInstance.Guid

	body := fmt.Sprintf(`{
		"type": "app",
		"relationships": {
			"app": {"guid" : "%s"},
			"service_instance": {"guid": "%s"}
		},
		"data": {
			"parameters": %s
		}
	}`, appGuid, serviceInstanceGuid, customParameters)

	if _, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/service_bindings"), "-X", "POST", "-d", body); err != nil {
		fmt.Printf("Failed to bind app %s to service instance %s\n", appName, serviceInstanceName)
		return
	}

	fmt.Println("OK")
}
