package sync

import (
	"context"
	"sync"
)

type FakeAdvertiser struct {
	mu          sync.Mutex
	advertising bool
	hostname    string
	port        int
	err         error
}

func NewFakeAdvertiser() *FakeAdvertiser {
	return &FakeAdvertiser{}
}

func (a *FakeAdvertiser) SetError(err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.err = err
}

func (a *FakeAdvertiser) Advertise(_ context.Context, hostname string, port int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.err != nil {
		return a.err
	}
	a.advertising = true
	a.hostname = hostname
	a.port = port
	return nil
}

func (a *FakeAdvertiser) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.advertising = false
}

func (a *FakeAdvertiser) IsAdvertising() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.advertising
}

func (a *FakeAdvertiser) Hostname() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.hostname
}

func (a *FakeAdvertiser) Port() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.port
}

type FakeBrowser struct {
	mu     sync.Mutex
	offers []PairingOffer
	err    error
}

func NewFakeBrowser() *FakeBrowser {
	return &FakeBrowser{}
}

func (b *FakeBrowser) SetOffers(offers []PairingOffer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.offers = offers
}

func (b *FakeBrowser) SetError(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.err = err
}

func (b *FakeBrowser) Browse(_ context.Context) (<-chan PairingOffer, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.err != nil {
		return nil, b.err
	}
	ch := make(chan PairingOffer, len(b.offers))
	for _, o := range b.offers {
		ch <- o
	}
	close(ch)
	return ch, nil
}
