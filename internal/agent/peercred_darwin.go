//go:build darwin

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

	var cred *unix.Xucred
	var credErr error

	err = raw.Control(func(fd uintptr) {
		cred, credErr = unix.GetsockoptXucred(int(fd), unix.SOL_LOCAL, unix.LOCAL_PEERCRED)
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

	var pid int
	var pidErr error

	err = raw.Control(func(fd uintptr) {
		pid, pidErr = unix.GetsockoptInt(int(fd), unix.SOL_LOCAL, unix.LOCAL_PEERPID)
	})
	if err != nil {
		return 0, fmt.Errorf("peercred: control failed: %w", err)
	}
	if pidErr != nil {
		return 0, fmt.Errorf("peercred: getsockopt failed: %w", pidErr)
	}

	return int32(pid), nil
}
