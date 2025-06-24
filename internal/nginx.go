package internal

import (
	"fmt"
	"github.com/spf13/afero"
	"os/exec"
	"path/filepath"
	"os"
)

type ProxyServer interface {
	WriteServer(server Server) error
	ReadServers() ([]Server, error)
	RemoveServer(server Server) error
	Restart() error
}

type Nginx struct {
	fs afero.Fs
	sitesPath string
	certsPath string
}

func NewNginx(fs afero.Fs, certsPath, sitesPath string) *Nginx {
	return &Nginx{
		fs: fs,
		sitesPath: sitesPath,
		certsPath: certsPath,
	}
}

func (nginx *Nginx) WriteServer(server Server) error {
	crtPath := filepath.Join(nginx.certsPath, server.Domain+".crt")
	keyPath := filepath.Join(nginx.certsPath, server.Domain+".key")
	if err := afero.WriteFile(nginx.fs, crtPath, []byte(server.Cert.Certificate), 0600); err != nil {
		return err
	}
	if err := afero.WriteFile(nginx.fs, keyPath, []byte(server.Cert.PrivateKey), 0600); err != nil {
		return err
	}

	path := filepath.Join(nginx.sitesPath, server.Domain)
	if err := afero.WriteFile(nginx.fs, path, []byte(server.Snippet), 0644); err != nil {
		return err
	}

	return nil
}

func (nginx *Nginx) ReadServers() ([]Server, error) {
	var servers []Server

	files, err := os.ReadDir(nginx.sitesPath)
	if err != nil {
		fmt.Printf("Failed to read server settings dir: %v\n", err)
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		domain := file.Name()
		snippetPath := filepath.Join(nginx.sitesPath, domain)
		snippetBytes, err := os.ReadFile(snippetPath)
		if err != nil {
			fmt.Printf("Failed to read snippet for %s: %v\n", domain, err)
			continue
		}

		crtPath := filepath.Join(nginx.certsPath, fmt.Sprintf("%s.crt", domain))
		keyPath := filepath.Join(nginx.certsPath, fmt.Sprintf("%s.key", domain))
		crtBytes, err := os.ReadFile(crtPath)
		if err != nil {
			fmt.Printf("Failed to read cert for %s: %v\n", domain, err)
			continue
		}
		keyBytes, err := os.ReadFile(keyPath)
		if err != nil {
			fmt.Printf("Failed to read key for %s: %v\n", domain, err)
			continue
		}

		server := Server{
			Domain:   domain,
			Snippet: string(snippetBytes),
			Cert: Cert{
				Certificate: string(crtBytes),
				PrivateKey:  string(keyBytes),
			},
		}
		servers = append(servers, server)
	}

	return servers, nil
}

func (nginx *Nginx) RemoveServer(server Server) error {
	path := filepath.Join(nginx.sitesPath, server.Domain)
	if err := nginx.fs.Remove(path); err != nil {
		return err
	}

	return nil
}

func (nginx *Nginx) Restart() error {
	cmd := exec.Command("nginx", "-s", "reload")
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	} else {
		return nil
	}
}
