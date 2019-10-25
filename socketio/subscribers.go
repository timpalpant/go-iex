package socketio

import "sync"

// Enables subscribing based on string symbols.
type Subscriber interface {
	// Subscribes to the given symbol.
	Subscribe(symbol string)
	// Returns true if the given symbol is currently subscribed to.
	Subscribed(symbol string) bool
	// Unsubscribed from events for the given symbol.
	Unsubscribe(symbol string)
}

// A Subscriber implementation using simple map presence.
type PresenceSubscriber struct {
	// Guards the symbols map.
	sync.RWMutex

	// Stores subscribed sympols.
	symbols map[string]bool
}

func (p *PresenceSubscriber) Subscribe(symbol string) {
	p.Lock()
	defer p.Unlock()
	p.symbols[symbol] = true
}

func (p *PresenceSubscriber) Subscribed(symbol string) bool {
	p.RLock()
	defer p.RUnlock()
	_, ok := p.symbols[symbol]
	return ok
}

func (p *PresenceSubscriber) Unsubscribe(symbol string) {
	p.Lock()
	defer p.Unlock()
	delete(p.symbols, symbol)
}

func NewPresenceSubscriber() Subscriber {
	return &PresenceSubscriber{symbols: make(map[string]bool)}
}

// A subscriber implementation using a counter. A certain number of Subscribe
// calls for a given symbol must be followed by the same number of Unsubscribe
// calls for the same symbol for Unsubscribed to return false.
type CountingSubscriber struct {
	// Guards the symbols map.
	sync.RWMutex

	// Stores subscribed sympols.
	symbols map[string]int
}

func (c *CountingSubscriber) Subscribe(symbol string) {
	c.Lock()
	defer c.Unlock()
	c.symbols[symbol]++
}

func (c *CountingSubscriber) Subscribed(symbol string) bool {
	c.RLock()
	defer c.RUnlock()
	return c.symbols[symbol] > 0
}

func (c *CountingSubscriber) Unsubscribe(symbol string) {
	c.Lock()
	defer c.Unlock()
	if c.symbols[symbol] > 0 {
		c.symbols[symbol]--
	} else {
		delete(c.symbols, symbol)
	}
}

func NewCountingSubscriber() Subscriber {
	return &CountingSubscriber{symbols: make(map[string]int)}
}
