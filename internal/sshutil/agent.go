package sshutil

import (
	"fmt"
	"net"

	"github.com/niklucky/wombat/internal/platform"
	"golang.org/x/crypto/ssh/agent"
)

// GetAgent returns an ssh-agent client if available.
func GetAgent() (agent.Agent, error) {
	socket := platform.AgentSocketPath()
	if socket == "" {
		return nil, fmt.Errorf("SSH_AUTH_SOCK not set")
	}
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, err
	}
	return agent.NewClient(conn), nil
}
