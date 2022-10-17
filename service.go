// Package emutex provides mutual exclusion for the group of entities.
// The mutex remains locked unless all the appropriate entities are unlocked.
package emutex

import (
	"golang.org/x/exp/constraints"
	"sync"
)

// Service describes the library interface.
type Service[T constraints.Ordered] interface {
	Lock(ids []T)
	Unlock(ids []T)
}

// NewService creates a new service instance for the specific entity's type.
func NewService[T constraints.Ordered]() *EntityMutex[T] {
	var condMux sync.Mutex
	return &EntityMutex[T]{
		unlockCondMux: &condMux,
		unlockCond:    sync.NewCond(&condMux),
		entityMux:     &sync.Mutex{},
		entityLock:    make(map[T]bool),
	}
}

// EntityMutex is a service implementation.
type EntityMutex[T constraints.Ordered] struct {
	// used for syncing the sync.Cond
	unlockCondMux *sync.Mutex
	// used for getting an update that indicates unlocked entities
	unlockCond *sync.Cond

	// used for locking/unlocking entities
	entityMux *sync.Mutex
	// storage of the locked/unlocked entities
	entityLock map[T]bool
}

// Lock locks the group of entities.
func (s EntityMutex[T]) Lock(ids []T) {
	// check if the required entities are unlocked
	if s.tryLock(ids) {
		// if yes then lock them and proceed
		return
	}
	// if we are here then at least one required entity is locked
	for {
		// waiting for update that should be sent when at least one entity is unlocked
		s.unlockCondMux.Lock()
		s.unlockCond.Wait()
		s.unlockCondMux.Unlock()
		// got an update
		// check if the required entities are unlocked
		if s.tryLock(ids) {
			// if yes then lock them and proceed
			return
		}
		// if no, then wait for another update
	}
}

// Unlock unlocks the group of entities.
func (s EntityMutex[T]) Unlock(ids []T) {
	s.entityMux.Lock()
	// unlock the required entities
	for _, id := range ids {
		s.entityLock[id] = false
	}
	s.entityMux.Unlock()
	// send an update that indicates that some entities were unlocked
	s.unlockCond.Broadcast()
}

// isAvailable checks if all required entities are unlocked
func (s EntityMutex[T]) isAvailable(ids []T) bool {
	var locked bool
	for _, id := range ids {
		if s.entityLock[id] {
			locked = true
			break
		}
	}
	return !locked
}

// lock marks all required entities as locked
func (s EntityMutex[T]) tryLock(ids []T) (success bool) {
	s.entityMux.Lock()
	defer s.entityMux.Unlock()
	// check if the required entities are unlocked
	if s.isAvailable(ids) {
		// if yes then lock them
		for _, id := range ids {
			s.entityLock[id] = true
		}
		success = true
	}
	return
}
