package commands

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/cf/ui_helpers"
	consumer "github.com/cloudfoundry/loggregator_consumer"
	"github.com/cloudfoundry/loggregatorlib/logmessage"

	"github.com/cloudfoundry/cli/plugin"
	. "github.com/cloudfoundry/v3-cli-plugin/models"
	. "github.com/cloudfoundry/v3-cli-plugin/util"
)

func Logs(cliConnection plugin.CliConnection, args []string) {
	appName := args[1]

	bufferTime := 5 * time.Second
	messageQueue := api.NewLoggregator_SortedMessageQueue(bufferTime, time.Now)

	onConnect := func() {
		fmt.Printf("Tailing logs for app %s...\r\n\r\n", appName)
	}

	onMessage := func(msg *logmessage.LogMessage) {
		fmt.Printf("%s\r\n", logMessageOutput(msg, time.Local))
	}

	output, _ := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps?names=%s", appName))
	apps := V3AppsModel{}
	json.Unmarshal([]byte(output[0]), &apps)

	if len(apps.Apps) == 0 {
		fmt.Printf("App %s not found\n", appName)
		return
	}
	app := apps.Apps[0]

	ssl, err := cliConnection.IsSSLDisabled()
	FreakOut(err)
	loggregatorEndpoint, err := cliConnection.LoggregatorEndpoint()
	FreakOut(err)
	accessToken, err := cliConnection.AccessToken()
	FreakOut(err)

	tlsConfig := net.NewTLSConfig([]tls.Certificate{}, ssl)
	loggregatorConsumer := consumer.New(loggregatorEndpoint, tlsConfig, http.ProxyFromEnvironment)
	defer func() {
		loggregatorConsumer.Close()
		flushMessageQueue(onMessage, messageQueue)
	}()

	loggregatorConsumer.SetOnConnectCallback(onConnect)
	logChan, err := loggregatorConsumer.Tail(app.Guid, accessToken)
	if err != nil {
		FreakOut(err)
	}

	bufferMessages(logChan, onMessage, messageQueue)
}

func bufferMessages(logChan <-chan *logmessage.LogMessage, onMessage func(*logmessage.LogMessage), messageQueue *api.Loggregator_SortedMessageQueue) {

	for {
		sendMessages(messageQueue, onMessage)

		select {
		case msg, ok := <-logChan:
			if !ok {
				return
			}
			messageQueue.PushMessage(msg)
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func flushMessageQueue(onMessage func(*logmessage.LogMessage), messageQueue *api.Loggregator_SortedMessageQueue) {
	if onMessage == nil {
		return
	}

	for {
		message := messageQueue.PopMessage()
		if message == nil {
			break
		}

		onMessage(message)
	}

	onMessage = nil
}

func sendMessages(queue *api.Loggregator_SortedMessageQueue, onMessage func(*logmessage.LogMessage)) {
	for queue.NextTimestamp() < time.Now().UnixNano() {
		msg := queue.PopMessage()
		onMessage(msg)
	}
}

func logMessageOutput(msg *logmessage.LogMessage, loc *time.Location) string {
	logHeader, coloredLogHeader := ui_helpers.ExtractLogHeader(msg, loc)
	logContent := ui_helpers.ExtractLogContent(msg, logHeader)

	return fmt.Sprintf("%s%s", coloredLogHeader, logContent)
}
