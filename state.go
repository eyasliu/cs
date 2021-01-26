package cs

import (
	"strings"
	"time"

	"github.com/gogf/gf/os/gcache"
)

// State 会话的状态数据管理
type State struct {
	cache            *gcache.Cache
	keyExpireTimeout time.Duration
}

// Get 获取指定会话的指定 key 的状态值
func (s *State) Get(sid, key string) interface{} {
	ck := s.getCacheKey(sid, key)
	v, _ := s.cache.Get(ck)
	return v
}

// Set 设置指定会话的状态键值对
func (s *State) Set(sid, key string, val interface{}) {
	ck := s.getCacheKey(sid, key)
	s.cache.Set(ck, val, s.keyExpireTimeout)
}

// 销毁指定会话的所有状态
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

// SetAdapter 设置会话状态的存储适配器，参考 goframe 的缓存管理适配器
// See: https://itician.org/pages/viewpage.action?pageId=1114265
func (s *State) SetAdapter(a gcache.Adapter) {
	s.cache.SetAdapter(a)
}

func (s *State) getCacheKey(sid, key string) string {
	return sid + ":" + key
}
