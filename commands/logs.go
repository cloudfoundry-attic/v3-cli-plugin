package commands

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	api "github.com/cloudfoundry/cli/cf/api/logs"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/cf/uihelpers"
	consumer "github.com/cloudfoundry/loggregator_consumer"
	"github.com/cloudfoundry/loggregatorlib/logmessage"

	"github.com/cloudfoundry/cli/plugin"
	. "github.com/cloudfoundry/v3-cli-plugin/models"
	. "github.com/cloudfoundry/v3-cli-plugin/util"
)

func Logs(cliConnection plugin.CliConnection, args []string) {
	appName := args[1]
	output, _ := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps?names=%s", appName))
	apps := V3AppsModel{}
	json.Unmarshal([]byte(output[0]), &apps)

	if len(apps.Apps) == 0 {
		fmt.Printf("App %s not found\n", appName)
		return
	}
	app := apps.Apps[0]

	messageQueue := api.NewLoggregatorMessageQueue()

	bufferTime := 25 * time.Millisecond
	ticker := time.NewTicker(bufferTime)

	c := make(chan *logmessage.LogMessage)

	loggregatorEndpoint, err := cliConnection.LoggregatorEndpoint()
	FreakOut(err)

	ssl, err := cliConnection.IsSSLDisabled()
	FreakOut(err)
	tlsConfig := net.NewTLSConfig([]tls.Certificate{}, ssl)

	loggregatorConsumer := consumer.New(loggregatorEndpoint, tlsConfig, http.ProxyFromEnvironment)
	defer func() {
		loggregatorConsumer.Close()
		flushMessageQueue(c, messageQueue)
	}()

	onConnect := func() {
		fmt.Printf("Tailing logs for app %s...\r\n\r\n", appName)
	}
	loggregatorConsumer.SetOnConnectCallback(onConnect)

	accessToken, err := cliConnection.AccessToken()
	FreakOut(err)

	logChan, err := loggregatorConsumer.Tail(app.Guid, accessToken)
	if err != nil {
		FreakOut(err)
	}

	go func() {
		for _ = range ticker.C {
			flushMessageQueue(c, messageQueue)
		}
	}()

	go func() {
		for msg := range logChan {
			messageQueue.PushMessage(msg)
		}

		flushMessageQueue(c, messageQueue)
		close(c)
	}()

	for msg := range c {
		fmt.Printf("%s\r\n", logMessageOutput(msg, time.Local))
	}
}

func flushMessageQueue(c chan *logmessage.LogMessage, messageQueue *api.LoggregatorMessageQueue) {
	messageQueue.EnumerateAndClear(func(m *logmessage.LogMessage) {
		c <- m
	})
}

func logMessageOutput(msg *logmessage.LogMessage, loc *time.Location) string {
	logHeader, coloredLogHeader := uihelpers.ExtractLogHeader(msg, loc)
	logContent := uihelpers.ExtractLogContent(msg, logHeader)

	return fmt.Sprintf("%s%s", coloredLogHeader, logContent)
}
