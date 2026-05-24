package tunnelmgr

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/niklucky/wombat/internal/core"
	"github.com/niklucky/wombat/internal/sshutil"
	"golang.org/x/crypto/ssh"
)

// Manager manages active SSH tunnels.
type Manager struct {
	mu            sync.RWMutex
	clients       map[string]*ssh.Client
	listeners     map[string]net.Listener
	quit          map[string]chan struct{}
	wg            map[string]*sync.WaitGroup
	keepaliveStop map[string]func()
}

// NewManager creates a new tunnel manager.
func NewManager() *Manager {
	return &Manager{
		clients:       make(map[string]*ssh.Client),
		listeners:     make(map[string]net.Listener),
		quit:          make(map[string]chan struct{}),
		wg:            make(map[string]*sync.WaitGroup),
		keepaliveStop: make(map[string]func()),
	}
}

// Start begins forwarding a tunnel. It dials SSH, listens locally, and forwards connections.
func (m *Manager) Start(tunnel core.Tunnel, host core.Host) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.clients[tunnel.Name]; ok {
		return fmt.Errorf("tunnel %q already active", tunnel.Name)
	}

	client, err := sshutil.Dial(host)
	if err != nil {
		return fmt.Errorf("ssh dial: %w", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", tunnel.LocalPort))
	if err != nil {
		client.Close()
		return fmt.Errorf("listen localhost:%d: %w", tunnel.LocalPort, err)
	}

	m.clients[tunnel.Name] = client
	m.listeners[tunnel.Name] = listener
	m.quit[tunnel.Name] = make(chan struct{})
	m.keepaliveStop[tunnel.Name] = sshutil.StartKeepalive(client, 30*time.Second)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	m.wg[tunnel.Name] = wg

	go func() {
		defer wg.Done()
		m.serve(tunnel, client, listener, m.quit[tunnel.Name])
	}()
	return nil
}

// Stop closes an active tunnel and waits for the serve goroutine to fully exit.
func (m *Manager) Stop(name string) error {
	m.mu.Lock()

	quit, ok := m.quit[name]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("tunnel %q not active", name)
	}
	close(quit)

	if stop, ok := m.keepaliveStop[name]; ok {
		stop()
	}
	if l, ok := m.listeners[name]; ok {
		l.Close()
	}
	if c, ok := m.clients[name]; ok {
		c.Close()
	}

	wg := m.wg[name]
	delete(m.clients, name)
	delete(m.listeners, name)
	delete(m.quit, name)
	delete(m.wg, name)
	delete(m.keepaliveStop, name)
	m.mu.Unlock()

	if wg != nil {
		wg.Wait()
	}
	return nil
}

// Status returns a map of tunnel names to active state.
func (m *Manager) Status() map[string]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]bool)
	for name := range m.clients {
		status[name] = true
	}
	return status
}

// IsActive reports whether a tunnel is currently running.
func (m *Manager) IsActive(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.clients[name]
	return ok
}

func (m *Manager) serve(tunnel core.Tunnel, client *ssh.Client, listener net.Listener, quit chan struct{}) {
	for {
		select {
		case <-quit:
			return
		default:
		}

		localConn, err := listener.Accept()
		if err != nil {
			select {
			case <-quit:
				return
			default:
				continue
			}
		}

		go m.forward(localConn, client, tunnel.RemoteHost, tunnel.RemotePort)
	}
}

func (m *Manager) forward(localConn net.Conn, client *ssh.Client, remoteHost string, remotePort int) {
	defer localConn.Close()

	remoteAddr := fmt.Sprintf("%s:%d", remoteHost, remotePort)
	remoteConn, err := client.Dial("tcp", remoteAddr)
	if err != nil {
		return
	}
	defer remoteConn.Close()

	done := make(chan struct{}, 2)
	go func() {
		io.Copy(remoteConn, localConn)
		done <- struct{}{}
	}()
	go func() {
		io.Copy(localConn, remoteConn)
		done <- struct{}{}
	}()
	<-done
}
