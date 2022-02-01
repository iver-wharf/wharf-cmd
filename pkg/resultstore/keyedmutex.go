package resultstore

import "sync"

type keyedMutex struct {
	m sync.Map
}

func (k keyedMutex) Lock(key interface{}) {
	val, _ := k.m.LoadOrStore(key, &sync.Mutex{})
	mutex := val.(*sync.Mutex)
	mutex.Lock()
}

func (k keyedMutex) Unlock(key interface{}) {
	val, ok := k.m.Load(key)
	if !ok {
		return
	}
	mutex := val.(*sync.Mutex)
	mutex.Unlock()
}
