package agent

const (
	ActionEncrypt  = "encrypt"
	ActionDecrypt  = "decrypt"
	ActionPing     = "ping"
	ActionShutdown = "shutdown"

	ProtocolVersion = 1
)

type Request struct {
	Version int    `json:"version"`
	Action  string `json:"action"`
	Payload []byte `json:"payload,omitempty"`
	Salt    []byte `json:"salt,omitempty"`
	Nonce   []byte `json:"nonce,omitempty"`
}

type Response struct {
	OK      bool   `json:"ok"`
	Payload []byte `json:"payload,omitempty"`
	Salt    []byte `json:"salt,omitempty"`
	Nonce   []byte `json:"nonce,omitempty"`
	Error   string `json:"error,omitempty"`
}
