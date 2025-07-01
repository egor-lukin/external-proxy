package internal

import (
	"reflect"
	"testing"
)

type mockProxyServer struct {
	servers        []Server
	writtenServers []Server
	removedServers []Server
	restartCalled  bool
	readErr        error
	writeErr       error
	removeErr      error
	restartErr     error
}

func (m *mockProxyServer) WriteServer(server Server) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.writtenServers = append(m.writtenServers, server)
	return nil
}
func (m *mockProxyServer) ReadServers() ([]Server, error) {
	return m.servers, m.readErr
}
func (m *mockProxyServer) RemoveServer(server Server) error {
	if m.removeErr != nil {
		return m.removeErr
	}
	m.removedServers = append(m.removedServers, server)
	return nil
}
func (m *mockProxyServer) Restart() error {
	m.restartCalled = true
	return m.restartErr
}

type mockKubeApi struct {
	servers []Server
	readErr error
}

func (m *mockKubeApi) ReadServers() ([]Server, error) {
	return m.servers, m.readErr
}

func TestSynchronizeServers(t *testing.T) {
	// local: a.com, b.com
	// remote: a.com (changed), c.com (new)
	local := []Server{
		{Domain: "a.com", Snippet: "foo", Cert: Cert{Certificate: "certA", PrivateKey: "keyA"}},
		{Domain: "b.com", Snippet: "bar", Cert: Cert{Certificate: "certB", PrivateKey: "keyB"}},
	}
	remote := []Server{
		{Domain: "a.com", Snippet: "foo2 ({{ .CertsPath }})", Cert: Cert{Certificate: "certA2", PrivateKey: "keyA2"}}, // changed
		{Domain: "c.com", Snippet: "baz", Cert: Cert{Certificate: "certC", PrivateKey: "keyC"}},                       // new
	}

	mockNginx := &mockProxyServer{servers: local}
	mockKube := &mockKubeApi{servers: remote}

	SynchronizeServers(mockNginx, mockKube, "certs_path")

	expectedWritten := []Server{
		{Domain: "a.com", Snippet: "foo2 (certs_path)", Cert: Cert{Certificate: "certA2", PrivateKey: "keyA2"}},
		{Domain: "c.com", Snippet: "baz", Cert: Cert{Certificate: "certC", PrivateKey: "keyC"}},
	}
	expectedRemoved := []Server{
		{Domain: "b.com", Snippet: "bar", Cert: Cert{Certificate: "certB", PrivateKey: "keyB"}},
	}

	if !reflect.DeepEqual(mockNginx.writtenServers, expectedWritten) {
		t.Errorf("Written servers mismatch: got %v, want %v", mockNginx.writtenServers, expectedWritten)
	}
	if !reflect.DeepEqual(mockNginx.removedServers, expectedRemoved) {
		t.Errorf("Removed servers mismatch: got %v, want %v", mockNginx.removedServers, expectedRemoved)
	}
	if !mockNginx.restartCalled {
		t.Errorf("Expected Restart to be called")
	}
}
