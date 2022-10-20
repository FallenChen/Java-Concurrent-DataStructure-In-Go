package concurrentmap

import (
	"encoding/json"
	"sync"
)

var SHARD_COUNT = 32

// A "thread" safe map of type string:Anything
// To avoid lock bottlenecks this map is dived to several (SHARD_COUNT) map shards
type ConcurrentMap[V any] []*ConcurrentMapShared[V]

type ConcurrentMapShared[V any] struct {
	items        map[string]V
	sync.RWMutex // Read Write mutex, guards access to internal map
}

func New[V any]() ConcurrentMap[V] {
	m := make(ConcurrentMap[V], SHARD_COUNT)
	for i := 0; i < SHARD_COUNT; i++ {
		m[i] = &ConcurrentMapShared[V]{items: make(map[string]V)}
	}
	return m
}

func (m ConcurrentMap[V]) GetShard(key string) *ConcurrentMapShared[V] {
	return m[uint(fnv32(key))%uint(SHARD_COUNT)]
}

func (m ConcurrentMap[V]) MSet(data map[string]V) {
	for key, value := range data {
		shard := m.GetShard(key)
		shard.Lock()
		shard.items[key] = value
		shard.Unlock()
	}
}

func (m ConcurrentMap[V]) Set(key string, value V) {
	shard := m.GetShard(key)
	shard.Lock()
	shard.items[key] = value
	shard.Unlock()
}

type UpsertCb[V any] func(exist bool, valueInMap V, newValue V) V

// Insert or Update - updates existing element or inserts a new one using UpsertCb
func (m ConcurrentMap[V]) Upsert(key string, value V, cb UpsertCb[V]) (res V) {
	shard := m.GetShard(key)
	shard.Lock()
	v, ok := shard.items[key]
	res = cb(ok, v, value)
	shard.items[key] = res
	shard.Unlock()
	return res
}

func (m ConcurrentMap[V]) SetIfAbsent(key string, value V) bool {

	shard := m.GetShard(key)
	shard.Lock()
	_, ok := shard.items[key]
	if !ok {
		shard.items[key] = value
	}
	shard.Unlock()
	return !ok
}

func (m ConcurrentMap[V]) Get(key string) (V, bool) {
	shard := m.GetShard(key)
	shard.RLock()
	// Get item from shard
	val, ok := shard.items[key]
	shard.RUnlock()
	return val, ok
}

func (m ConcurrentMap[V]) Count() int {
	count := 0
	for i := 0; i < SHARD_COUNT; i++ {
		shard := m[i]
		shard.RLock()
		count += len(shard.items)
		shard.RUnlock()
	}
	return count
}

// Looks up an item under specified key
func (m ConcurrentMap[V]) Has(key string) bool {

	shard := m.GetShard(key)
	shard.RLock()
	_, ok := shard.items[key]
	shard.RUnlock()
	return ok
}

func (m ConcurrentMap[V]) Remove(key string) {
	shard := m.GetShard(key)
	shard.Lock()
	delete(shard.items, key)
	shard.Unlock()
}

// RemoveCb is a callback executed in a map.RemoveCb() call, while Lock is held
// It reutrns true, the element will be removed from the map
type RemoveCb[V any] func(key string, v V, exists bool) bool

func (m ConcurrentMap[V]) RemoveCb(key string, cb RemoveCb[V]) bool {
	shard := m.GetShard(key)
	shard.Lock()
	v, ok := shard.items[key]
	remove := cb(key, v, ok)
	if remove && ok {
		delete(shard.items, key)
	}
	shard.Unlock()
	return remove
}

// Pop removes an element from the map and returns it
func (m ConcurrentMap[V]) Pop(key string) (v V, exists bool) {
	shard := m.GetShard(key)
	shard.Lock()
	v, exists = shard.items[key]
	delete(shard.items, key)
	shard.Unlock()
	return v, exists
}

func (m ConcurrentMap[V]) IsEmpty() bool {
	return m.Count() == 0
}

// Used by the IterBuffered functions to wrap two variables together over a channel
type Tuple[V any] struct {
	Key string
	Val V
}

func (m ConcurrentMap[V]) Iter() <-chan Tuple[V] {
	chans := snapshot(m)
	ch := make(chan Tuple[V])
	go fanIn(chans, ch)
	return ch
}

// IterBuffered returns an iterator which could bu used in a for range loop
func (m ConcurrentMap[V]) IterBuffered() <-chan Tuple[V] {
	chans := snapshot(m)
	total := 0
	for _, c := range chans {
		total += cap(c)
	}
	ch := make(chan Tuple[V], total)
	go fanIn(chans, ch)
	return ch
}

func (m ConcurrentMap[V]) Clear() {
	for item := range m.IterBuffered() {
		m.Remove(item.Key)
	}
}

// Returns a array of channels that contains elements in each shard,
// which likely takes a snapshot of 'm'
// It returns once the size of each buffered channel is determined,
// before all the channels are populated using goroutines
func snapshot[V any](m ConcurrentMap[V]) (chans []chan Tuple[V]) {
	if len(m) == 0 {
		panic(`cmap.ConcurrentMap is not initialized. Should run New() before usage.`)
	}

	chans = make([]chan Tuple[V], SHARD_COUNT)
	wg := sync.WaitGroup{}
	wg.Add(SHARD_COUNT)
	// Foreach shard
	for index, shard := range m {
		go func(index int, shard *ConcurrentMapShared[V]) {
			// Foreach key, value pair
			shard.RLock()
			chans[index] = make(chan Tuple[V], len(shard.items))
			wg.Done()
			for key, val := range shard.items {
				chans[index] <- Tuple[V]{key, val}
			}
			shard.RUnlock()
			close(chans[index])
		}(index, shard)
	}
	wg.Wait()
	return chans
}

// fanIn reads elements from channels `chans` into channel `out`
func fanIn[V any](chans []chan Tuple[V], out chan Tuple[V]) {
	wg := sync.WaitGroup{}
	wg.Add(len(chans))
	for _, ch := range chans {
		go func(ch chan Tuple[V]) {
			for t := range ch {
				out <- t
			}
			wg.Done()
		}(ch)
	}
	wg.Wait()
	close(out)
}

func (m ConcurrentMap[V]) Items() map[string]V {
	tmp := make(map[string]V)

	// Insert items to temporary map.
	for item := range m.IterBuffered() {
		tmp[item.Key] = item.Val
	}

	return tmp
}

type IterCb[V any] func(key string, v V)

func (m ConcurrentMap[V]) IterCb(fn IterCb[V]) {
	for idx := range m {
		shard := (m)[idx]
		shard.RLock()
		for key, value := range shard.items {
			fn(key, value)
		}
		shard.RUnlock()
	}
}

func (m ConcurrentMap[V]) Keys() []string {
	count := m.Count()
	ch := make(chan string, count)
	go func() {
		// Foreach shard.
		wg := sync.WaitGroup{}
		wg.Add(SHARD_COUNT)
		for _, shard := range m {
			go func(shard *ConcurrentMapShared[V]) {
				// Foreach key, value pair.
				shard.RLock()
				for key := range shard.items {
					ch <- key
				}
				shard.RUnlock()
				wg.Done()
			}(shard)
		}
		wg.Wait()
		close(ch)
	}()

	// Generate keys
	keys := make([]string, 0, count)
	for k := range ch {
		keys = append(keys, k)
	}
	return keys
}

func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	keyLength := len(key)
	for i := 0; i < keyLength; i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}

func (m ConcurrentMap[V]) MarshalJSON() ([]byte, error) {
	tmp := make(map[string]V)

	for item := range m.IterBuffered() {
		tmp[item.Key] = item.Val
	}
	return json.Marshal(tmp)
}

func (m *ConcurrentMap[V]) UnmarshalJSON(b []byte) (err error) {
	tmp := make(map[string]V)

	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}

	for key, val := range tmp {
		m.Set(key, val)
	}
	return nil
}
