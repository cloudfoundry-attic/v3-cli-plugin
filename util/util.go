package util

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/plugin"
)

func Poll(cliConnection plugin.CliConnection, endpoint string, desired string, timeout time.Duration, timeoutMessage string) {
	timeElapsed := 0 * time.Second
	for timeElapsed < timeout {
		rawOutput, err := cliConnection.CliCommandWithoutTerminalOutput("curl", endpoint, "-X", "GET")
		FreakOut(err)
		output := strings.Join(rawOutput, "")
		if strings.Contains(output, desired) {
			return
		}
		timeElapsed = timeElapsed + 1*time.Second
		time.Sleep(1 * time.Second)
	}
	FreakOut(errors.New(timeoutMessage))
}
func FreakOut(err error) {
	if err != nil {
		fmt.Println("Error Will Robinson!: ", err.Error())
		os.Exit(1)
	}

}
