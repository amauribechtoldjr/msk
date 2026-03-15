//go:build windows

package agent

import "net"

// verifyPeerCredentials is a no-op on Windows.
// Windows AF_UNIX sockets do not support peer credential retrieval.
func verifyPeerCredentials(conn net.Conn) error {
	return nil
}

// getPeerPID is a no-op on Windows.
// Windows AF_UNIX sockets do not support peer PID retrieval.
func getPeerPID(conn net.Conn) (int32, error) {
	return 0, nil
}
