package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var CouponCache *cache.Cache

func Init() {
	// 10 min default expiration, 15 min cleanup interval
	CouponCache = cache.New(10*time.Minute, 15*time.Minute)
}
