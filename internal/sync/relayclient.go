package sync

import "github.com/fnune/kyaraben/internal/syncthing"

type RelayClient = syncthing.RelayClient
type CreateSessionResponse = syncthing.CreateSessionResponse
type GetSessionResponse = syncthing.GetSessionResponse
type GetResponseResponse = syncthing.GetResponseResponse

var ProductionRelayURLs = syncthing.ProductionRelayURLs

func NewRelayClient(urls []string) (*RelayClient, error) {
	return syncthing.NewRelayClient(urls)
}

func NewDefaultRelayClient() (*RelayClient, error) {
	return syncthing.NewDefaultRelayClient()
}
