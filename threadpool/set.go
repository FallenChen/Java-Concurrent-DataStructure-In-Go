package threadpool

import "sync"

type Set struct {
	_map *sync.Map
}

func NewSet() *Set {
	set := new(Set)
	set._map = new(sync.Map)
	return set
}

func (s *Set) Add(value interface{}) {
	s._map.Store(value, true)
}

func (s *Set) Remove(value interface{}) {
	s._map.Delete(value)
}

func (s *Set) Contains(value interface{}) bool {
	_, ok := s._map.Load(value)
	return ok
}

func (s *Set) GetAll() []interface{} {
	values := make([]interface{},0)
	s._map.Range(func(key interface{}, value interface{}) bool {
		values = append(values, key)
		return true
	})
	return values
}