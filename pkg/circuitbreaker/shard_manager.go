package circuitbreaker

import "sync"

type ShardManager struct {
	mu       sync.Mutex
	breakers map[string]*CircuitBreaker
	cfg      Config
}

func NewShardManager(cfg Config) *ShardManager {
	return &ShardManager{
		breakers: make(map[string]*CircuitBreaker),
		cfg:      cfg,
	}
}

func (m *ShardManager) ForShard(shardURL string) *CircuitBreaker {
	m.mu.Lock()
	defer m.mu.Unlock()

	breaker, ok := m.breakers[shardURL]
	if ok {
		return breaker
	}

	breaker = NewCircuitBreaker(m.cfg)
	m.breakers[shardURL] = breaker
	return breaker
}
