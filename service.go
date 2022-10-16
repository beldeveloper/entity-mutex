package emutex

import (
	"golang.org/x/exp/constraints"
	"sync"
)

type Service[T constraints.Ordered] interface {
	Lock(ids []T)
	Unlock(ids []T)
}

func NewService[T constraints.Ordered]() *EntityMutex[T] {
	var condMux sync.Mutex
	return &EntityMutex[T]{
		unlockUpdateMux: &condMux,
		unlockUpdate:    sync.NewCond(&condMux),
		entityMux:       &sync.Mutex{},
		entityLock:      make(map[T]bool),
	}
}

type EntityMutex[T constraints.Ordered] struct {
	unlockUpdateMux *sync.Mutex
	unlockUpdate    *sync.Cond

	entityMux  *sync.Mutex
	entityLock map[T]bool
}

func (s EntityMutex[T]) Lock(ids []T) {
	s.entityMux.Lock()
	if s.isAvailable(ids) {
		s.lock(ids)
		s.entityMux.Unlock()
		return
	}
	s.entityMux.Unlock()
	for {
		s.unlockUpdateMux.Lock()
		s.unlockUpdate.Wait()
		s.unlockUpdateMux.Unlock()
		s.entityMux.Lock()
		if s.isAvailable(ids) {
			s.lock(ids)
			s.entityMux.Unlock()
			return
		}
		s.entityMux.Unlock()
	}
}

func (s EntityMutex[T]) Unlock(ids []T) {
	s.entityMux.Lock()
	for _, id := range ids {
		s.entityLock[id] = false
	}
	s.entityMux.Unlock()
	s.unlockUpdate.Broadcast()
}

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

func (s EntityMutex[T]) lock(ids []T) {
	for _, id := range ids {
		s.entityLock[id] = true
	}
}
