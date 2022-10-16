package emutex

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestService(t *testing.T) {
	partsN := 10
	partSize := 50
	doneExp := int64(partsN * partSize)
	var done int64

	service := NewService[int]()

	for i := 0; i < partsN; i++ {
		var wg sync.WaitGroup
		wg.Add(partSize)
		for j := 0; j < partSize; j++ {
			id1 := getRandID()
			id2 := getRandID()
			ids := make([]int, 0, 2)
			ids = append(ids, id1)
			if id1 != id2 {
				ids = append(ids, id2)
			}
			go someImportantAction(service, &wg, &done, ids, getRandDuration())
		}
		wg.Wait()
	}

	if doneExp != done {
		t.Errorf("invalid number of executed functions; expected: %d; done %d", doneExp, done)
	}
}

func getRandID() int {
	rand.Seed(time.Now().UnixNano())
	min := 1
	max := 10
	return rand.Intn(max-min+1) + min
}

func getRandDuration() time.Duration {
	rand.Seed(time.Now().UnixNano())
	min := 10
	max := 50
	n := rand.Intn(max-min+1) + min
	return time.Millisecond * time.Duration(n)
}

func someImportantAction(s Service[int], wg *sync.WaitGroup, done *int64, ids []int, workTime time.Duration) {
	s.Lock(ids)
	time.Sleep(workTime)
	s.Unlock(ids)
	wg.Done()
	atomic.AddInt64(done, 1)
}
