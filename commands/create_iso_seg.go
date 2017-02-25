package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
	. "github.com/cloudfoundry/v3-cli-plugin/models"
)

func CreateIsolationSegment(cliConnection plugin.CliConnection, args []string) {
	isoSegName := args[1]

	fmt.Printf("Creating isolation segment %s...\n", isoSegName)

	body := fmt.Sprintf(`{
		"name": "%s"
	}`, isoSegName)

	rawOutput, _ := cliConnection.CliCommandWithoutTerminalOutput("curl", "/v3/isolation_segments", "-X", "POST", "-d", body)
	isoSeg := V3IsolationSegmentModel{}
	output := strings.Join(rawOutput, "")
	json.Unmarshal([]byte(output), &isoSeg)

	if isoSeg.Name != isoSegName || isoSeg.Guid == "" {
		fmt.Printf("Failed to create isolation segment %s\n", isoSegName)
		return
	}

	fmt.Println("OK")
}
