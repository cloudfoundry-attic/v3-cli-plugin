package commands

import (
	"fmt"

	"github.com/jberkhahn/cli/plugin"
)

func Apps(cliConnection plugin.CliConnection, args []string) {
	fmt.Println("apps!")
	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", "v3/apps", "-X", "GET")

}
