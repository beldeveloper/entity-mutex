# Entity Mutex

## Description
The library provides mutual exclusion for the group of entities. The mutex remains locked unless all the appropriate entities are unlocked.

## Requirements
```text
Go >= 1.18
```

## Installation
```shell
go get -u github.com/beldeveloper/entity-mutex
```

## Example

```go
package main

import emutex "github.com/beldeveloper/entity-mutex"

func main() {
	service := emutex.NewService[int]()
	// goroutine #1
	// it locks entities #1 and #2 unless the work is done
	go doWork(service, []int{1, 2})
	// goroutine #2
	// the lock initiated by goroutine #1 doesn't affect goroutine #2 because it operates by other entities
	go doWork(service, []int{3, 4})
	// goroutine #3
	// it will wait for goroutines #1 and #2 because they have locked entities #1 and #3
	go doWork(service, []int{1, 3})
}

func doWork(service emutex.Service[int], entities []int) {
	service.Lock(entities)
	// do something
	service.Unlock(entities)
}

```