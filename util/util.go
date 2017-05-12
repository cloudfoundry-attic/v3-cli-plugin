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
	PollWithBadString(cliConnection, endpoint, desired, "", timeout, timeoutMessage)
}

func PollWithBadString(cliConnection plugin.CliConnection, endpoint string, desired string, bad string, timeout time.Duration, timeoutMessage string) {
	timeElapsed := 0 * time.Second
	var output string
	for timeElapsed < timeout {
		rawOutput, err := cliConnection.CliCommandWithoutTerminalOutput("curl", endpoint, "-X", "GET")
		FreakOut(err)
		output = strings.Join(rawOutput, "")
		if strings.Contains(output, desired) {
			return
		}
		if bad != "" && strings.Contains(output, bad) {
			FreakOut(fmt.Errorf("output contains problem: [%s]: %s", bad, output))
		}
		timeElapsed = timeElapsed + 1*time.Second
		time.Sleep(1 * time.Second)
	}
	FreakOut(errors.New(timeoutMessage + "\n" + output))
}
func FreakOut(err error) {
	if err != nil {
		fmt.Println("Error Will Robinson!: ", err.Error())
		os.Exit(1)
	}

}
