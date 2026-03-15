package agent

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/amauribechtoldjr/msk/internal/gcm"
)

type Client struct {
	sockPath string
}

func NewClient(sockPath string) *Client {
	return &Client{sockPath: sockPath}
}

func (c *Client) send(req Request) (*Response, error) {
	conn, err := net.Dial("unix", c.sockPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agent: %w", err)
	}
	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	var resp Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if !resp.OK {
		return nil, fmt.Errorf("agent error: %s", resp.Error)
	}

	return &resp, nil
}

// Encrypt implements vault.Vault
func (c *Client) Encrypt(data []byte) (*gcm.SaltedGCM, error) {
	resp, err := c.send(Request{
		Version: ProtocolVersion,
		Action:  ActionEncrypt,
		Payload: data,
	})
	if err != nil {
		return nil, err
	}

	return &gcm.SaltedGCM{
		CipherData: resp.Payload,
		Salt:       resp.Salt,
		Nonce:      resp.Nonce,
	}, nil
}

// Decrypt implements vault.Vault
func (c *Client) Decrypt(salt, nonce, data []byte) ([]byte, error) {
	resp, err := c.send(Request{
		Version: ProtocolVersion,
		Action:  ActionDecrypt,
		Payload: data,
		Salt:    salt,
		Nonce:   nonce,
	})
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// DestroyMK implements vault.Vault — sends shutdown to agent
func (c *Client) DestroyMK() {
	c.send(Request{
		Version: ProtocolVersion,
		Action:  ActionShutdown,
	})
}

// LoadMK implements vault.Vault — no-op for agent client
func (c *Client) LoadMK() error {
	return nil
}

// Ping checks if the agent is alive
func (c *Client) Ping() error {
	_, err := c.send(Request{
		Version: ProtocolVersion,
		Action:  ActionPing,
	})
	return err
}
