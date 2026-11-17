package sse

import (
	"encoding/json"
	"sync"
)

type Broker struct {
	mu   sync.RWMutex
	subs map[string]map[chan []byte]struct{}
}

func New() *Broker { return &Broker{subs: map[string]map[chan []byte]struct{}{}} }

func (b *Broker) Subscribe(endpointID string) chan []byte {
	ch := make(chan []byte, 16)
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.subs[endpointID] == nil {
		b.subs[endpointID] = map[chan []byte]struct{}{}
	}
	b.subs[endpointID][ch] = struct{}{}
	return ch
}

func (b *Broker) Unsubscribe(endpointID string, ch chan []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if m, ok := b.subs[endpointID]; ok {
		delete(m, ch)
		if len(m) == 0 {
			delete(b.subs, endpointID)
		}
	}
	close(ch)
}

func (b *Broker) Publish(endpointID string, v any) {
	data, _ := json.Marshal(v)
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subs[endpointID] {
		select {
		case ch <- data:
		default:
		}
	}
}
