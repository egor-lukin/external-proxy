package internal

import (
	"text/template"
	"bytes"
)

type Cert struct {
	PrivateKey  string
	Certificate string
}

type Server struct {
	Domain   string
	Snippet  string
	Cert     Cert
}

func (s Server) FillTemplate(certsPath string) error {
	tmpl, err := template.New("nginx").Parse(s.Snippet)
	if err != nil {
		return err
	}

	data := struct{ CertsPath string }{CertsPath: certsPath}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	return nil
}

func (s Server) Equal(other Server) bool {
	return s.Cert.Certificate == other.Cert.Certificate &&
		s.Cert.PrivateKey == other.Cert.PrivateKey &&
		s.Snippet == other.Snippet
}

func SelectChangedServers(localServers, removeServers []Server) []Server {
	oldMap := make(map[string]Server)
	for _, s := range localServers {
		oldMap[s.Domain] = s
	}

	var filtered []Server
	for _, s := range removeServers {
		old, exists := oldMap[s.Domain]

		if !exists || !s.Equal(old) {
			filtered = append(filtered, s)
		}
	}

	return filtered
}

func SelectRemovedServers(localServers, remoteServers []Server) []Server {
	newDomains := make(map[string]struct{})
	for _, s := range remoteServers {
		newDomains[s.Domain] = struct{}{}
	}
	var removedServers []Server
	for _, s := range localServers {
		if _, exists := newDomains[s.Domain]; !exists {
			removedServers = append(removedServers, s)
		}
	}

	return removedServers
}
