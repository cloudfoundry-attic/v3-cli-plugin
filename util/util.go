package util

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jberkhahn/cli/plugin"
)

func Poll(cliConnection plugin.CliConnection, endpoint string, desired string, timeout time.Duration, timeoutMessage string) {
	timeElapsed := 0 * time.Second
	for timeElapsed < timeout {
		output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", endpoint, "-X", "GET")
		FreakOut(err)
		if strings.Contains(output[0], desired) {
			return
		}
		timeElapsed = timeElapsed + 1*time.Second
		time.Sleep(1 * time.Second)
	}
}
func FreakOut(err error) {
	if err != nil {
		fmt.Println("Error Will Robinson!: ", err.Error())
		os.Exit(1)
	}

}
