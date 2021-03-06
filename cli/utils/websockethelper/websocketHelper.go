package websockethelper

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/gorilla/websocket"
	"github.com/keptn/keptn/cli/utils"
)

// MyCloudEvent represents a keptn cloud event
type MyCloudEvent struct {
	CloudEventsVersion string          `json:"cloudEventsVersion"`
	ContentType        string          `json:"contentType"`
	Data               json.RawMessage `json:"data"`
	EventID            string          `json:"eventID"`
	EventTime          string          `json:"eventTime"`
	EventType          string          `json:"eventType"`
	Type               string          `json:"type"`
	Source             string          `json:"source"`
}

// LogData represents log data
type LogData struct {
	Message   string `json:"message"`
	Terminate bool   `json:"terminate"`
	LogLevel  string `json:"loglevel"`
}

// incompleteCE is a helper type for unmarshalling the CE data
type incompleteCE struct {
	ConnData ConnectionData `json:"data"`
}

// ConnectionData stores ChannelInfo and Success data
type ConnectionData struct {
	ChannelInfo ChannelInfo `json:"channelInfo"`
}

// ChannelInfo stores a token and a channelID used for opening the websocket
type ChannelInfo struct {
	Token     string `json:"token"`
	ChannelID string `json:"channelId"`
}

// PrintWSContentCEResponse opens a websocket using the passed
// connection data (in form of a cloud event) and prints status data
func PrintWSContentCEResponse(responseCE *cloudevents.Event, controlEndPoint url.URL) error {

	ceData := &incompleteCE{}
	err := responseCE.DataAs(ceData)
	if err != nil {
		return err
	}
	return printWSContent(ceData.ConnData, controlEndPoint)
}

// PrintWSContentByteResponse opens a websocket using the passed
// connection data (in form of a byte slice) and prints status data
func PrintWSContentByteResponse(response []byte, controlEndPoint url.URL) error {

	ceData := &incompleteCE{}
	err := json.Unmarshal(response, ceData)
	if err != nil {
		return err
	}

	return printWSContent(ceData.ConnData, controlEndPoint)
}

func printWSContent(connData ConnectionData, controlEndPoint url.URL) error {

	err := validateConnectionData(connData)
	if err != nil {
		return err
	}

	ws, _, err := openWS(connData, controlEndPoint)
	if err != nil {
		fmt.Println("Opening websocket failed")
		return err
	}
	// PrintLogLevel(LogData{Message: "Websocket successfully opened", LogLevel: "DEBUG"}, loglevel)
	defer ws.Close()

	return readAndPrintCE(ws)
}

func validateConnectionData(connData ConnectionData) error {
	if connData.ChannelInfo.Token == "" && connData.ChannelInfo.ChannelID == "" {
		return errors.New("Could not open websocket because Token or Channel ID are missing")
	}
	return nil
}

// openWS opens a websocket
func openWS(connData ConnectionData, controlEndPoint url.URL) (*websocket.Conn, *http.Response, error) {

	wsEndPoint := controlEndPoint
	wsEndPoint.Scheme = "wss"

	header := http.Header{}
	header.Add("Token", connData.ChannelInfo.Token)
	header.Add("x-keptn-ws-channel-id", connData.ChannelInfo.ChannelID)
	if strings.Contains(controlEndPoint.String(), "xip.io") {

		regex := `\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`
		re := regexp.MustCompile(regex)
		ip := re.FindString(controlEndPoint.String())

		header.Add("Host", controlEndPoint.Host)
		wsEndPoint.Host = ip
	}
	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	return dialer.Dial(wsEndPoint.String(), header)
}

// readAndPrintCE reads a cloud event from the websocket
func readAndPrintCE(ws *websocket.Conn) error {
	ws.SetReadDeadline(time.Now().Add(time.Minute))
	for {
		messageType, message, err := ws.ReadMessage()
		if messageType == 1 { // 1.. textmessage
			var messageCE MyCloudEvent
			dec := json.NewDecoder(strings.NewReader(string(message)))
			if err := dec.Decode(&messageCE); err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}
			if printCE(messageCE) {
				return nil
			}
		}
		if err != nil {
			log.Println("read: ", err)
			return err
		}

	}
	return nil
}

func printCE(ce MyCloudEvent) bool {
	var log LogData
	if err := json.Unmarshal(ce.Data, &log); err != nil {
		fmt.Println("JSON unmarshalling error. LogData format expected.")
		//return nil, err
	}
	switch ce.Type {
	case "sh.keptn.events.log":
		utils.PrintLogStringLevel(log.Message, log.LogLevel)
		return log.Terminate
	default:
		fmt.Println("type of event could not be processed")
	}
	return true
}
