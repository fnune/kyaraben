package sync

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/hashicorp/mdns"
)

const mdnsServiceType = "_kyaraben._tcp"

type MDNSAdvertiser struct {
	server *mdns.Server
}

func NewMDNSAdvertiser() *MDNSAdvertiser {
	return &MDNSAdvertiser{}
}

func (a *MDNSAdvertiser) Advertise(_ context.Context, hostname string, port int) error {
	ip := localIP()
	info := []string{fmt.Sprintf("pair=%s:%d", ip, port)}
	log.Info("Starting mDNS advertisement: %s on port %d (ip=%s)", hostname, port, ip)

	service, err := mdns.NewMDNSService(hostname, mdnsServiceType, "", "", port, nil, info)
	if err != nil {
		return fmt.Errorf("creating mDNS service: %w", err)
	}

	server, err := mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		return fmt.Errorf("starting mDNS server: %w", err)
	}

	a.server = server
	log.Info("mDNS advertisement started successfully")
	return nil
}

func (a *MDNSAdvertiser) Stop() {
	if a.server != nil {
		_ = a.server.Shutdown()
		a.server = nil
	}
}

type MDNSBrowser struct{}

func NewMDNSBrowser() *MDNSBrowser {
	return &MDNSBrowser{}
}

func (b *MDNSBrowser) Browse(ctx context.Context) (<-chan PairingOffer, error) {
	offers := make(chan PairingOffer, 16)
	entries := make(chan *mdns.ServiceEntry, 16)

	go func() {
		defer close(offers)
		for {
			select {
			case <-ctx.Done():
				return
			case entry, ok := <-entries:
				if !ok {
					return
				}
				offer := entryToOffer(entry)
				if offer != nil {
					select {
					case offers <- *offer:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	params := mdns.DefaultParams(mdnsServiceType)
	params.Entries = entries
	params.DisableIPv6 = true

	go func() {
		defer close(entries)
		_ = mdns.Query(params)
	}()

	return offers, nil
}

func entryToOffer(entry *mdns.ServiceEntry) *PairingOffer {
	for _, field := range entry.InfoFields {
		if strings.HasPrefix(field, "pair=") {
			return &PairingOffer{
				Hostname:    entry.Name,
				PairingAddr: strings.TrimPrefix(field, "pair="),
			}
		}
	}
	return nil
}

func localIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		name, _ := os.Hostname()
		return name
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	name, _ := os.Hostname()
	return name
}
