package tunnel

import "encoding/json"

// P2PRegisterPayload 节点向 gateway 注册 P2P peerID。
type P2PRegisterPayload struct {
	PeerID    string `json:"peer_id"`
	AuthToken string `json:"auth_token,omitempty"`
}

// DecodeP2PRegisterPayload 解析 P2P 注册负载。
func DecodeP2PRegisterPayload(data []byte) (P2PRegisterPayload, error) {
	var p P2PRegisterPayload
	if len(data) == 0 {
		return p, ErrInvalidMessage
	}
	err := json.Unmarshal(data, &p)
	return p, err
}
