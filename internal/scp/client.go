package scp

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Client struct {
	config    *ssh.ClientConfig
	port      int
	recursive bool
}

func NewClient(user, password string, port int, recursive bool) *Client {
	var authMethods []ssh.AuthMethod

	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	} else {
		// 尝试加载 SSH 密钥
		home, err := os.UserHomeDir()
		if err == nil {
			key, err := os.ReadFile(filepath.Join(home, ".ssh", "id_rsa"))
			if err == nil {
				signer, err := ssh.ParsePrivateKey(key)
				if err == nil {
					authMethods = append(authMethods, ssh.PublicKeys(signer))
				}
			}
		}
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: authMethods,
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		}),
		Timeout: 30 * time.Second,
	}

	return &Client{
		config:    config,
		port:      port,
		recursive: recursive,
	}
}

func (c *Client) connect(host string) (*ssh.Client, error) {
	// 如果 host 包含用户名（格式如 user@host），需要提取出真实的 host
	if idx := strings.Index(host, "@"); idx != -1 {
		host = host[idx+1:]
	}
	return ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, c.port), c.config)
}
