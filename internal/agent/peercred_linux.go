//go:build linux

package agent

import (
	"fmt"
	"net"
	"os"

	"golang.org/x/sys/unix"
)

func verifyPeerCredentials(conn net.Conn) error {
	uc, ok := conn.(*net.UnixConn)
	if !ok {
		return fmt.Errorf("peercred: not a unix connection")
	}

	raw, err := uc.SyscallConn()
	if err != nil {
		return fmt.Errorf("peercred: failed to get raw conn: %w", err)
	}

	var cred *unix.Ucred
	var credErr error

	err = raw.Control(func(fd uintptr) {
		cred, credErr = unix.GetsockoptUcred(int(fd), unix.SOL_SOCKET, unix.SO_PEERCRED)
	})
	if err != nil {
		return fmt.Errorf("peercred: control failed: %w", err)
	}
	if credErr != nil {
		return fmt.Errorf("peercred: getsockopt failed: %w", credErr)
	}

	if cred.Uid != uint32(os.Getuid()) {
		return fmt.Errorf("peercred: uid mismatch: peer=%d self=%d", cred.Uid, os.Getuid())
	}

	return nil
}

func getPeerPID(conn net.Conn) (int32, error) {
	uc, ok := conn.(*net.UnixConn)
	if !ok {
		return 0, fmt.Errorf("peercred: not a unix connection")
	}

	raw, err := uc.SyscallConn()
	if err != nil {
		return 0, fmt.Errorf("peercred: failed to get raw conn: %w", err)
	}

	var cred *unix.Ucred
	var credErr error

	err = raw.Control(func(fd uintptr) {
		cred, credErr = unix.GetsockoptUcred(int(fd), unix.SOL_SOCKET, unix.SO_PEERCRED)
	})
	if err != nil {
		return 0, fmt.Errorf("peercred: control failed: %w", err)
	}
	if credErr != nil {
		return 0, fmt.Errorf("peercred: getsockopt failed: %w", credErr)
	}

	return cred.Pid, nil
}
