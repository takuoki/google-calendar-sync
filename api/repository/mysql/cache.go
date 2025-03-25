package mysql

import (
	"sync"

	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
)

var refreshTokenCache = newRefreshTokenCache()

type RefreshTokenCache struct {
	mu       sync.RWMutex
	tokenMap map[valueobject.CalendarID]string
}

func newRefreshTokenCache() *RefreshTokenCache {
	return &RefreshTokenCache{
		tokenMap: make(map[valueobject.CalendarID]string),
	}
}

func (c *RefreshTokenCache) Get(calendarID valueobject.CalendarID) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	token, ok := c.tokenMap[calendarID]
	return token, ok
}

func (c *RefreshTokenCache) Set(calendarID valueobject.CalendarID, token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokenMap[calendarID] = token
}

func (c *RefreshTokenCache) Delete(calendarID valueobject.CalendarID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.tokenMap, calendarID)
}
