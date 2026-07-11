//go:build linux || darwin

package main

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

func dropPrivileges() error {
	if os.Getuid() != 0 {
		return nil
	}

	u, err := user.Lookup("nobody")
	if err != nil {
		return fmt.Errorf("cant find user nobody: %w", err)
	}

	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(u.Gid)

	if err = syscall.Setgid(gid); err != nil {
		return fmt.Errorf("setgid failed: %w", err)
	}

	if err = syscall.Setuid(uid); err != nil {
		return fmt.Errorf("setuid failed: %w", err)
	}

	return nil
}
