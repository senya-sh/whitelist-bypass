package pion

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

type SignalingMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
	ID   int             `json:"id,omitempty"`
	Role string          `json:"role,omitempty"`
}

type ICEServerConfig struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

type SDPMessage struct {
	Type string `json:"type"`
	SDP  string `json:"sdp"`
}

type ICECandidateMessage struct {
	Candidate     string `json:"candidate"`
	SDPMid        string `json:"sdpMid"`
	SDPMLineIndex uint16 `json:"sdpMLineIndex"`
}

var WsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func ParseICEServers(data json.RawMessage) ([]webrtc.ICEServer, error) {
	var servers []ICEServerConfig
	if err := json.Unmarshal(data, &servers); err != nil {
		return nil, err
	}
	iceServers := make([]webrtc.ICEServer, len(servers))
	for i, s := range servers {
		iceServers[i] = webrtc.ICEServer{
			URLs: s.URLs, Username: s.Username, Credential: s.Credential,
		}
	}
	return iceServers, nil
}

func NewPionAPI(localIP string) *webrtc.API {
	se := webrtc.SettingEngine{}
	se.SetNet(&AndroidNet{LocalIP: localIP})
	return webrtc.NewAPI(webrtc.WithSettingEngine(se))
}

type WSHelper struct {
	wsConn *websocket.Conn
	mu     sync.Mutex
}

func (h *WSHelper) SetConn(ws *websocket.Conn) {
	h.mu.Lock()
	h.wsConn = ws
	h.mu.Unlock()
}

func (h *WSHelper) SendToHook(msgType string, data any) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.wsConn == nil {
		return
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return
	}
	msg := SignalingMessage{Type: msgType, Data: dataBytes}
	msgBytes, _ := json.Marshal(msg)
	h.wsConn.WriteMessage(websocket.TextMessage, msgBytes)
}

func (h *WSHelper) SendToHookWithRole(msgType string, data any, role string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.wsConn == nil {
		return
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return
	}
	msg := SignalingMessage{Type: msgType, Data: dataBytes, Role: role}
	msgBytes, _ := json.Marshal(msg)
	h.wsConn.WriteMessage(websocket.TextMessage, msgBytes)
}

func (h *WSHelper) SendResponse(id int, data any) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.wsConn == nil {
		return
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return
	}
	msg := SignalingMessage{Type: "response", Data: dataBytes, ID: id}
	msgBytes, _ := json.Marshal(msg)
	h.wsConn.WriteMessage(websocket.TextMessage, msgBytes)
}

func (h *WSHelper) ReadMessages(handler func([]byte), onDisconnect func()) {
	for {
		_, msg, err := h.wsConn.ReadMessage()
		if err != nil {
			onDisconnect()
			return
		}
		handler(msg)
	}
}
