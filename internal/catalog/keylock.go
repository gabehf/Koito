package catalog

import (
	"hash/fnv"
	"sync"
)

// keyLocks is a lock map for preventing duplicate entires from being creates.
var keyLocks [256]sync.Mutex

// lockKey locks the stripe for key and returns a function to unlock it.
func lockKey(key string) func() {
	h := fnv.New32a()
	h.Write([]byte(key))
	mu := &keyLocks[h.Sum32()%uint32(len(keyLocks))]
	mu.Lock()
	return mu.Unlock
}
