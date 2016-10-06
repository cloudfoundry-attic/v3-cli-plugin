package commands

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/cf/api/logs"
	"github.com/cloudfoundry/cli/cf/net"
	consumer "github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"

	"github.com/cloudfoundry/cli/plugin"
	. "github.com/cloudfoundry/v3-cli-plugin/models"
	. "github.com/cloudfoundry/v3-cli-plugin/util"
)

func Logs(cliConnection plugin.CliConnection, args []string) {
	appName := args[1]
	rawOutput, _ := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps?names=%s", appName))
	apps := V3AppsModel{}
	output := strings.Join(rawOutput, "")
	json.Unmarshal([]byte(output), &apps)

	if len(apps.Apps) == 0 {
		fmt.Printf("App %s not found\n", appName)
		return
	}
	app := apps.Apps[0]

	messageQueue := logs.NewNoaaMessageQueue()

	bufferTime := 25 * time.Millisecond
	ticker := time.NewTicker(bufferTime)

	logChan := make(chan logs.Loggable)
	errChan := make(chan error)

	dopplerEndpoint, err := cliConnection.DopplerEndpoint()
	FreakOut(err)

	ssl, err := cliConnection.IsSSLDisabled()
	FreakOut(err)
	tlsConfig := net.NewTLSConfig([]tls.Certificate{}, ssl)

	noaaConsumer := consumer.New(dopplerEndpoint, tlsConfig, http.ProxyFromEnvironment)
	defer func() {
		noaaConsumer.Close()
		flushMessages(logChan, messageQueue)
	}()

	onConnect := func() {
		fmt.Printf("Tailing logs for app %s...\r\n\r\n", appName)
	}
	noaaConsumer.SetOnConnectCallback(onConnect)

	accessToken, err := cliConnection.AccessToken()
	FreakOut(err)

	c, e := noaaConsumer.TailingLogs(app.Guid, accessToken)

	go func() {
		for {
			select {
			case msg, ok := <-c:
				if !ok {
					ticker.Stop()
					flushMessages(logChan, messageQueue)
					close(logChan)
					close(errChan)
					return
				}

				messageQueue.PushMessage(msg)
			case err := <-e:
				if err != nil {
					errChan <- err

					ticker.Stop()
					close(logChan)
					close(errChan)
					return
				}
			}
		}
	}()

	go func() {
		for range ticker.C {
			flushMessages(logChan, messageQueue)
		}
	}()

	for {
		select {
		case msg := <-logChan:
			fmt.Printf("%s\r\n", logMessageOutput(msg, time.Local))
		case err, ok := <-errChan:
			if !ok {
				FreakOut(err)
			}
		}
	}
}

func flushMessages(c chan <- logs.Loggable, messageQueue *logs.NoaaMessageQueue) {
	messageQueue.EnumerateAndClear(func(m *events.LogMessage) {
		c <- logs.NewNoaaLogMessage(m)
	})
}
func logMessageOutput(msg logs.Loggable, loc *time.Location) string {
	return fmt.Sprintf("%s", msg.ToLog(loc))
}
