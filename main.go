package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/net/proxy"
)

// Message represents the SSAP message
type Message struct {
	Type    string      `json:"type"`
	ID      string      `json:"id,omitempty"`
	URI     string      `json:"uri,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
	Error   string      `json:"error,omitempty"`
}

var (
	addr        = flag.String("addr", "192.168.1.237:3000", "TV address")
	keyFile     = flag.String("key-file", "key", "Path to file containing Client Key")
	cmd         = flag.String("cmd", "", "Command: initialize-key, info, list-apps, launch, close, vol-get, vol-set, vol-up, vol-down, mute, un-mute, chan-get, chan-up, chan-down, toast, turn-off, list-inputs, set-input, play, pause, stop, rewind, fast-forward")
	arg         = flag.String("arg", "", "Argument for command")
	payload     = flag.String("payload", "", "Optional JSON payload for launch command")
	socks5Proxy = flag.String("use-socks5-proxy", "", "SOCKS5 proxy address (e.g., 127.0.0.1:1080)")
)

var clientKey string

func main() {
	log.SetOutput(os.Stderr)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nCommand Examples:")
		fmt.Fprintln(os.Stderr, "  initialize-key: -cmd initialize-key")
		fmt.Fprintln(os.Stderr, "  info:           -cmd info")
		fmt.Fprintln(os.Stderr, "  list-apps:      -cmd list-apps")
		fmt.Fprintln(os.Stderr, "  launch:         -cmd launch -arg youtube")
		fmt.Fprintln(os.Stderr, "  launch params:  -cmd launch -arg youtube -payload '{\"contentId\":\"...\"}'")
		fmt.Fprintln(os.Stderr, "  close:          -cmd close -arg youtube")
		fmt.Fprintln(os.Stderr, "  vol-get:        -cmd vol-get")
		fmt.Fprintln(os.Stderr, "  vol-set:        -cmd vol-set -arg 20")
		fmt.Fprintln(os.Stderr, "  vol-up:         -cmd vol-up")
		fmt.Fprintln(os.Stderr, "  vol-down:       -cmd vol-down")
		fmt.Fprintln(os.Stderr, "  mute:           -cmd mute")
		fmt.Fprintln(os.Stderr, "  un-mute:        -cmd un-mute")
		fmt.Fprintln(os.Stderr, "  chan-get:       -cmd chan-get")
		fmt.Fprintln(os.Stderr, "  chan-up:        -cmd chan-up")
		fmt.Fprintln(os.Stderr, "  chan-down:      -cmd chan-down")
		fmt.Fprintln(os.Stderr, "  toast:          -cmd toast -arg \"Hello World\"")
		fmt.Fprintln(os.Stderr, "  turn-off:       -cmd turn-off")
		fmt.Fprintln(os.Stderr, "  list-inputs:    -cmd list-inputs")
		fmt.Fprintln(os.Stderr, "  set-input:      -cmd set-input -arg HDMI_1")
		fmt.Fprintln(os.Stderr, "  media control:  -cmd play (pause, stop, rewind, fast-forward)")
	}
	flag.Parse()

	if *cmd == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Load key from file
	if content, err := ioutil.ReadFile(*keyFile); err == nil {
		clientKey = strings.TrimSpace(string(content))
	}

	if *cmd == "initialize-key" {
		clientKey = ""
	}

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
	log.Printf("Connecting to %s", u.String())

	var dialer *websocket.Dialer
	if *socks5Proxy != "" {
		log.Printf("Using SOCKS5 proxy: %s", *socks5Proxy)
		d := *websocket.DefaultDialer // Copy the struct
		d.NetDial = func(network, addr string) (net.Conn, error) {
			p, err := proxy.SOCKS5("tcp", *socks5Proxy, nil, proxy.Direct)
			if err != nil {
				return nil, err
			}
			return p.Dial(network, addr)
		}
		dialer = &d
	} else {
		dialer = websocket.DefaultDialer
	}

	c, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Context for multi-step requests
	reqContext := make(map[string]string)

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			var msgResp map[string]interface{}
			if err := json.Unmarshal(message, &msgResp); err != nil {
				continue
			}

			msgType, _ := msgResp["type"].(string)

			if msgType == "registered" {
				log.Println("Registered successfully!")
				payload, ok := msgResp["payload"].(map[string]interface{})
				if ok {
					if newKey, ok := payload["client-key"].(string); ok {
						if newKey != clientKey {
							clientKey = newKey
							log.Printf("New Client Key: %s", clientKey)
							_ = ioutil.WriteFile(*keyFile, []byte(clientKey), 0644)
						}
					}
				}
				// Execute the command
				executeCommand(c, *cmd, *arg, reqContext)

			} else if msgType == "response" {
				msgID, _ := msgResp["id"].(string)

				if strings.HasPrefix(msgID, "req_") {
					handleResponse(c, msgID, msgResp, *cmd, *arg, reqContext)
				}
			} else if msgType == "error" {
				log.Printf("Error: %v", msgResp)
				os.Exit(1)
			}
		}
	}()

	// Handshake
	handshake := map[string]interface{}{
		"type": "register",
		"id":   "register_0",
		"payload": map[string]interface{}{
			"forcePairing": false,
			"pairingType":  "PROMPT",
			"client-key":   clientKey,
			"manifest": map[string]interface{}{
				"manifestVersion": 1,
				"appVersion":      "1.1",
				"permissions": []string{
					"READ_UPDATE_INFO",
					"READ_NETWORK_STATE",
					"READ_RUNNING_APPS",
					"READ_INSTALLED_APPS",
					"CONTROL_AUDIO",
					"CONTROL_INPUT_TEXT",
					"CONTROL_MOUSE",
					"CONTROL_POWER",
					"CONTROL_TV",
					"READ_APP_STATUS",
					"READ_CURRENT_CHANNEL",
					"READ_INPUT_DEVICE_LIST",
					"READ_TV_CHANNEL_LIST",
					"READ_VOLUME",
					"WRITE_NOTIFICATION_TOAST",
					"LAUNCH",
					"CONTROL_APP",
					"WEBAPP_LAUNCHER",
					"CONTROL_INPUT_MEDIA_PLAYBACK",
					"CHECK_BLUETOOTH_DEVICE",
					"CONTROL_BLUETOOTH",
					"READ_SETTINGS",
					"CONTROL_DISPLAY",
					"READ_LGE_SDX",
					"READ_NOTIFICATIONS",
					"WRITE_SETTINGS",
					"TEST_SECURE",
					"CONTROL_MOUSE_AND_KEYBOARD",
				},
			},
		},
	}
	if err := c.WriteJSON(handshake); err != nil {
		log.Fatal("write handshake:", err)
	}

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func sendRequest(c *websocket.Conn, uri string, payload interface{}, id string) {
	req := Message{
		Type:    "request",
		ID:      id,
		URI:     uri,
		Payload: payload,
	}
	if err := c.WriteJSON(req); err != nil {
		log.Fatal("write request:", err)
	}
}

func executeCommand(c *websocket.Conn, command string, argument string, reqContext map[string]string) {
	switch command {
	case "info":
		sendRequest(c, "ssap://system/getSystemInfo", nil, "req_info")
	case "vol-get":
		sendRequest(c, "ssap://audio/getVolume", nil, "req_vol_get")
	case "vol-set":
		vol, _ := strconv.Atoi(argument)
		sendRequest(c, "ssap://audio/setVolume", map[string]interface{}{"volume": vol}, "req_vol_set")
	case "vol-up":
		sendRequest(c, "ssap://audio/volumeUp", nil, "req_vol_up")
	case "vol-down":
		sendRequest(c, "ssap://audio/volumeDown", nil, "req_vol_down")
	case "mute":
		sendRequest(c, "ssap://audio/setMute", map[string]interface{}{"mute": true}, "req_mute")
	case "un-mute":
		sendRequest(c, "ssap://audio/setMute", map[string]interface{}{"mute": false}, "req_unmute")
	case "chan-get":
		sendRequest(c, "ssap://tv/getCurrentChannel", nil, "req_chan_get")
	case "chan-up":
		sendRequest(c, "ssap://tv/channelUp", nil, "req_chan_up")
	case "chan-down":
		sendRequest(c, "ssap://tv/channelDown", nil, "req_chan_down")
	case "list-apps":
		sendRequest(c, "ssap://com.webos.applicationManager/listApps", nil, "req_list_apps")
	case "launch", "close":
		// Resolve app if needed
		if !strings.Contains(argument, ".") && argument != "" {
			reqContext["pending_cmd"] = command
			reqContext["pending_arg"] = argument
			sendRequest(c, "ssap://com.webos.applicationManager/listApps", nil, "req_resolve_app")
		} else {
			doAppAction(c, command, argument)
		}
	case "initialize-key":
		log.Println("Key initialized and saved to file.")
		os.Exit(0)
	case "toast":
		if argument == "" {
			log.Fatal("Toast message argument required")
		}
		sendRequest(c, "ssap://system.notifications/createToast", map[string]interface{}{"message": argument}, "req_toast")
	case "turn-off":
		sendRequest(c, "ssap://system/turnOff", nil, "req_turn_off")
	case "list-inputs":
		sendRequest(c, "ssap://tv/getExternalInputList", nil, "req_list_inputs")
	case "set-input":
		if argument == "" {
			log.Fatal("Input ID argument required")
		}
		sendRequest(c, "ssap://tv/switchInput", map[string]interface{}{"inputId": argument}, "req_set_input")
	case "play":
		sendRequest(c, "ssap://media.controls/play", nil, "req_play")
	case "pause":
		sendRequest(c, "ssap://media.controls/pause", nil, "req_pause")
	case "stop":
		sendRequest(c, "ssap://media.controls/stop", nil, "req_stop")
	case "rewind":
		sendRequest(c, "ssap://media.controls/rewind", nil, "req_rewind")
	case "fast-forward":
		sendRequest(c, "ssap://media.controls/fastForward", nil, "req_fast_forward")
	default:
		log.Printf("Unknown command: %s", command)
		os.Exit(1)
	}
}

func doAppAction(c *websocket.Conn, command, appId string) {
	if command == "launch" {
		reqPayload := map[string]interface{}{"id": appId}
		if *payload != "" {
			var params map[string]interface{}
			if err := json.Unmarshal([]byte(*payload), &params); err != nil {
				log.Fatalf("Invalid JSON payload: %v", err)
			}
			reqPayload["params"] = params
		}
		sendRequest(c, "ssap://system.launcher/launch", reqPayload, "req_launch")
	} else if command == "close" {
		sendRequest(c, "ssap://system.launcher/close", map[string]interface{}{"id": appId}, "req_close")
	}
}

func handleResponse(c *websocket.Conn, msgID string, msgResp map[string]interface{}, command string, argument string, reqContext map[string]string) {
	payload, _ := msgResp["payload"]

	if msgID == "req_resolve_app" {
		pendingCmd := reqContext["pending_cmd"]
		pendingArg := reqContext["pending_arg"]

		appsList, ok := payload.(map[string]interface{})["apps"].([]interface{})
		if !ok {
			log.Println("Could not parse app list for resolution")
			os.Exit(1)
		}

		foundID := ""
		for _, a := range appsList {
			appMap := a.(map[string]interface{})
			title, _ := appMap["title"].(string)
			id, _ := appMap["id"].(string)

			if strings.EqualFold(title, pendingArg) || strings.EqualFold(id, pendingArg) {
				foundID = id
				log.Printf("Resolved '%s' to ID: %s", pendingArg, foundID)
				break
			}
		}

		if foundID != "" {
			doAppAction(c, pendingCmd, foundID)
		} else {
			log.Printf("Could not find app with name: %s", pendingArg)
			os.Exit(1)
		}
		return
	}

	if strings.HasSuffix(msgID, "_get") || msgID == "req_info" || msgID == "req_list_apps" || msgID == "req_list_inputs" {
		jsonOut, _ := json.MarshalIndent(payload, "", "  ")
		fmt.Fprintln(os.Stdout, string(jsonOut))
	} else {
		log.Printf("Command %s request sent/acknowledged.", command)
	}

	os.Exit(0)
}
