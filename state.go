package cmdsrv

import (
	"strings"
	"time"

	"github.com/gogf/gf/os/gcache"
)

// State 会话的状态数据管理
type State struct {
	cache *gcache.Cache
}

const keyExpireTimeout = 24 * time.Hour

func (s *State) Get(sid, key string) interface{} {
	ck := s.getCacheKey(sid, key)
	v, _ := s.cache.Get(ck)
	return v
}
func (s *State) Set(sid, key string, val interface{}) {
	ck := s.getCacheKey(sid, key)
	s.cache.Set(ck, val, keyExpireTimeout)
}

func (s *State) destroySid(sid string) {
	keys, err := s.cache.Keys()
	if err != nil {
		return
	}
	prefix := sid + ":"
	ks := []interface{}{}
	for _, ck := range keys {
		if k, ok := ck.(string); !ok || strings.Index(k, prefix) == 0 {
			ks = append(ks, ck)
		}
	}
	s.cache.Removes(ks)
}

func (s *State) SetAdapter(a gcache.Adapter) {
	s.cache.SetAdapter(a)
}
func (s *State) getCacheKey(sid, key string) string {
	return sid + ":" + key
}
