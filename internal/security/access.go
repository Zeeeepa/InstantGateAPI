package security

import (
	"sync"

	"github.com/proyaai/instantgate/internal/config"
)

type AccessControl struct {
	cfg          *config.SecurityConfig
	whitelistMap map[string]bool
	blacklistMap map[string]bool
	mu           sync.RWMutex
}

func NewAccessControl(cfg *config.SecurityConfig) *AccessControl {
	ac := &AccessControl{
		cfg:          cfg,
		whitelistMap: make(map[string]bool),
		blacklistMap: make(map[string]bool),
	}

	for _, table := range cfg.Whitelist {
		ac.whitelistMap[table] = true
	}

	for _, table := range cfg.Blacklist {
		ac.blacklistMap[table] = true
	}

	return ac
}

func (ac *AccessControl) IsTableAllowed(table string) bool {
	if !ac.cfg.Enabled {
		return true
	}

	ac.mu.RLock()
	defer ac.mu.RUnlock()

	if ac.blacklistMap[table] {
		return false
	}

	if len(ac.whitelistMap) == 0 {
		return true
	}

	return ac.whitelistMap[table]
}

func (ac *AccessControl) AddToWhitelist(table string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.whitelistMap[table] = true
}

func (ac *AccessControl) AddToBlacklist(table string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.blacklistMap[table] = true
}

func (ac *AccessControl) RemoveFromWhitelist(table string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	delete(ac.whitelistMap, table)
}

func (ac *AccessControl) RemoveFromBlacklist(table string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	delete(ac.blacklistMap, table)
}

func (ac *AccessControl) GetWhitelist() []string {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	result := make([]string, 0, len(ac.whitelistMap))
	for table := range ac.whitelistMap {
		result = append(result, table)
	}
	return result
}

func (ac *AccessControl) GetBlacklist() []string {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	result := make([]string, 0, len(ac.blacklistMap))
	for table := range ac.blacklistMap {
		result = append(result, table)
	}
	return result
}
