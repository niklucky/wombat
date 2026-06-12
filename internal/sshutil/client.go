package sshutil

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/niklucky/wombat/internal/core"
	"github.com/niklucky/wombat/internal/locales"
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

	addr := net.JoinHostPort(host.Address, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, config.Timeout)
	if err != nil {
		return nil, err
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return ssh.NewClient(c, chans, reqs), nil
}

// StartKeepalive starts a background goroutine that sends SSH keepalive
// requests at the given interval. If a keepalive fails or times out, the
// client is closed so that callers can detect the broken connection.
// The returned stop function should be called to clean up the goroutine.
func StartKeepalive(client *ssh.Client, interval time.Duration) func() {
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := sendKeepalive(client); err != nil {
					client.Close()
					return
				}
			case <-stop:
				return
			}
		}
	}()
	return func() { close(stop) }
}

func sendKeepalive(client *ssh.Client) error {
	done := make(chan error, 1)
	go func() {
		_, _, err := client.SendRequest("keepalive@openssh.com", true, nil)
		done <- err
	}()
	select {
	case err := <-done:
		return err
	case <-time.After(10 * time.Second):
		return locales.Errorf("errors.keepaliveTimedOut")
	}
}

// TestConnection attempts a TCP dial to the target address and port.
func TestConnection(host core.Host) error {
	port := host.Port
	if port == 0 {
		port = 22
	}
	addr := net.JoinHostPort(host.Address, fmt.Sprintf("%d", port))
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
		return nil, locales.Errorf("errors.noAuthMethods")
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
