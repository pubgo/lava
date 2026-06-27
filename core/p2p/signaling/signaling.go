package signaling

import "context"

// Type 信令消息类型。
type Type uint8

const (
	TypeOffer Type = iota + 1
	TypeAnswer
	TypeCandidate
	TypeBye
)

// Message 在节点间交换的 ICE 信令（JSON 序列化，经 tunnel 控制流传输）。
type Message struct {
	Type      Type   `json:"type"`
	From      string `json:"from"`
	To        string `json:"to"`
	Ufrag     string `json:"ufrag,omitempty"`
	Pwd       string `json:"pwd,omitempty"`
	Candidate string `json:"candidate,omitempty"` // SDP 格式 candidate 行
	AuthToken string `json:"auth_token,omitempty"`
}

// Broker 抽象信令收发；tunnel 实现见 tunnelsig（P1 后续）。
type Broker interface {
	Send(ctx context.Context, msg Message) error
	Recv(ctx context.Context, selfID string) (Message, error)
	Close() error
}
