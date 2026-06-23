package models

// Host represents an SSH target.
type Host struct {
	Name      string `json:"name"`
	Address   string `json:"address"`
	User      string `json:"user"`
	Port      int    `json:"port"`
	KeyPath   string `json:"key_path"`
	ProxyJump string `json:"proxy_jump,omitempty"`
}

// Key represents an SSH key.
type Key struct {
	Path          string `json:"path"`
	Fingerprint   string `json:"fingerprint"`
	IsAgentLoaded bool   `json:"is_agent_loaded"`
}

// Tunnel represents an SSH port forwarding rule.
type Tunnel struct {
	Name       string `json:"name"`
	HostName   string `json:"host_name"`
	LocalPort  int    `json:"local_port"`
	RemoteHost string `json:"remote_host"`
	RemotePort int    `json:"remote_port"`
	Active     bool   `json:"active"`
}
