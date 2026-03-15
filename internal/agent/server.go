package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/amauribechtoldjr/msk/internal/vault"
)

type Server struct {
	vault    vault.Vault
	sockPath string
	listener net.Listener
	ttl      time.Duration
	timer    *time.Timer
	mu       sync.Mutex
	done     chan struct{}
}

func NewServer(masterPassword []byte, sockPath string, ttl time.Duration) (*Server, error) {
	v := vault.NewVaultWithMK(masterPassword)

	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on unix socket: %w", err)
	}

	if err := os.Chmod(sockPath, 0o600); err != nil {
		ln.Close()
		return nil, fmt.Errorf("failed to set socket permissions: %w", err)
	}

	return &Server{
		vault:    v,
		sockPath: sockPath,
		listener: ln,
		ttl:      ttl,
		done:     make(chan struct{}),
	}, nil
}

func (s *Server) Serve() {
	s.mu.Lock()
	s.timer = time.AfterFunc(s.ttl, s.Shutdown)
	s.mu.Unlock()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				return
			default:
				log.Printf("accept error: %v", err)
				continue
			}
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	if err := verifyPeerCredentials(conn); err != nil {
		log.Printf("peer credential verification failed: %v", err)
		return
	}

	pid, _ := getPeerPID(conn)
	log.Printf("client PID %d connected", pid)

	var req Request
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		resp := Response{OK: false, Error: "invalid request: " + err.Error()}
		json.NewEncoder(conn).Encode(resp)
		return
	}

	s.mu.Lock()
	if s.timer != nil {
		s.timer.Reset(s.ttl)
	}
	s.mu.Unlock()

	resp := s.handleRequest(req)
	json.NewEncoder(conn).Encode(resp)
}

func (s *Server) handleRequest(req Request) Response {
	switch req.Action {
	case ActionEncrypt:
		result, err := s.vault.Encrypt(req.Payload)
		if err != nil {
			return Response{OK: false, Error: err.Error()}
		}
		return Response{OK: true, Payload: result.CipherData, Salt: result.Salt, Nonce: result.Nonce}

	case ActionDecrypt:
		plaintext, err := s.vault.Decrypt(req.Salt, req.Nonce, req.Payload)
		if err != nil {
			return Response{OK: false, Error: err.Error()}
		}
		return Response{OK: true, Payload: plaintext}

	case ActionPing:
		return Response{OK: true}

	case ActionShutdown:
		go s.Shutdown()
		return Response{OK: true}

	default:
		return Response{OK: false, Error: fmt.Sprintf("unknown action: %s", req.Action)}
	}
}

func (s *Server) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.done:
		return
	default:
	}

	close(s.done)

	if s.timer != nil {
		s.timer.Stop()
	}

	s.listener.Close()
	os.Remove(s.sockPath)
	s.vault.DestroyMK()
}
