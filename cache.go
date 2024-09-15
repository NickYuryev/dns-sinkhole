package main

import (
	"fmt"
	"time"

	"github.com/miekg/dns"
	gocache "github.com/patrickmn/go-cache"
)

type Cache struct {
	c *gocache.Cache
}

func NewCache() *Cache {
	return &Cache{c: gocache.New(time.Minute, 30*time.Second)}
}

func cacheKey(q dns.Question) string {
	return fmt.Sprintf("%d.%d.%s", q.Qtype, q.Qclass, q.Name)
}

func (c *Cache) Get(q dns.Question) (*dns.Msg, bool) {
	start := time.Now()
	v, ok := c.c.Get(cacheKey(q))
	label := "miss"
	if ok {
		label = "hit"
	}
	cacheGetLatency.WithLabelValues(label).Observe(time.Since(start).Seconds())
	if !ok {
		return nil, false
	}
	return v.(*dns.Msg).Copy(), true
}

func (c *Cache) Set(q dns.Question, m *dns.Msg, ttl time.Duration) {
	start := time.Now()
	c.c.Set(cacheKey(q), m.Copy(), ttl)
	cacheSetLatency.Observe(time.Since(start).Seconds())
}
