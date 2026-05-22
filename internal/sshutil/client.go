package sshutil

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/niklucky/wombat/internal/core"
	"golang.org/x/crypto/ssh"
)

// Dial establishes an SSH connection to the given host.
func Dial(host core.Host) (*ssh.Client, error) {
	authMethods, err := collectAuthMethods(host)
	if err != nil {
		return nil, fmt.Errorf("auth methods: %w", err)
	}

	port := host.Port
	if port == 0 {
		port = 22
	}

	config := &ssh.ClientConfig{
		User:            host.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host.Address, port)
	return ssh.Dial("tcp", addr, config)
}

// TestConnection attempts a TCP dial to the target address and port.
func TestConnection(host core.Host) error {
	port := host.Port
	if port == 0 {
		port = 22
	}
	addr := fmt.Sprintf("%s:%d", host.Address, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return err
	}
	return conn.Close()
}

func collectAuthMethods(host core.Host) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	// 1. Explicit key path
	if host.KeyPath != "" {
		m, err := keyAuthMethod(host.KeyPath)
		if err == nil {
			methods = append(methods, m)
		}
	}

	// 2. Default keys
	home, _ := os.UserHomeDir()
	for _, name := range []string{"id_ed25519", "id_rsa", "id_ecdsa"} {
		path := filepath.Join(home, ".ssh", name)
		if _, err := os.Stat(path); err == nil {
			m, err := keyAuthMethod(path)
			if err == nil {
				methods = append(methods, m)
			}
		}
	}

	// 3. SSH agent
	if a, err := GetAgent(); err == nil {
		methods = append(methods, ssh.PublicKeysCallback(a.Signers))
	}

	if len(methods) == 0 {
		return nil, fmt.Errorf("no authentication methods available")
	}
	return methods, nil
}

func keyAuthMethod(path string) (ssh.AuthMethod, error) {
	key, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}
