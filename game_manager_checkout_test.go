package boardgame

import (
	"sync"
	"testing"
	"time"
)

type blockingGameLoadStorage struct {
	StorageManager
	entered chan struct{}
	release chan struct{}
}

func (s *blockingGameLoadStorage) Game(id string) (*GameStorageRecord, error) {
	s.entered <- struct{}{}
	<-s.release
	return s.StorageManager.Game(id)
}

func TestConcurrentColdCheckoutReturnsOneCanonicalGame(t *testing.T) {
	game := testDefaultGame(t, true)
	manager := game.Manager()
	manager.Internals().UseManualTimers()
	defer manager.Internals().Close()
	manager.freezeGame(game)

	storage := &blockingGameLoadStorage{
		StorageManager: manager.storage,
		entered:        make(chan struct{}, 2),
		release:        make(chan struct{}),
	}
	manager.storage = storage

	start := make(chan struct{})
	results := make(chan *Game, 2)
	var callers sync.WaitGroup
	callers.Add(2)
	for i := 0; i < 2; i++ {
		go func() {
			defer callers.Done()
			<-start
			results <- manager.ModifiableGame(game.ID())
		}()
	}
	close(start)

	select {
	case <-storage.entered:
	case <-time.After(time.Second):
		t.Fatal("first checkout did not reach storage")
	}
	// Give the other caller a chance to contend while the canonical load is
	// blocked. It must wait for that load rather than starting a second one.
	select {
	case <-storage.entered:
		close(storage.release)
		callers.Wait()
		t.Fatal("concurrent checkout started a duplicate storage load")
	case <-time.After(100 * time.Millisecond):
	}
	close(storage.release)
	callers.Wait()
	close(results)

	var canonical *Game
	for result := range results {
		if result == nil {
			t.Fatal("cold checkout returned nil")
		}
		if canonical == nil {
			canonical = result
		} else if result != canonical {
			t.Fatal("concurrent cold checkouts returned distinct mutable games")
		}
	}
	manager.modifiableGamesLock.RLock()
	resident := manager.modifiableGames[game.ID()]
	manager.modifiableGamesLock.RUnlock()
	if resident != canonical {
		t.Fatal("returned game was not the manager's canonical resident instance")
	}
}

func TestColdTimerAndRequestShareCanonicalGame(t *testing.T) {
	manager, game := durableTimerTestGame(t, newTestStorageManager())
	defer manager.Internals().Close()
	version := game.Version()
	manager.freezeGame(game)

	storage := &blockingGameLoadStorage{
		StorageManager: manager.storage,
		entered:        make(chan struct{}, 2),
		release:        make(chan struct{}),
	}
	manager.storage = storage

	type timerResult struct {
		fired bool
		err   error
	}
	start := make(chan struct{})
	requestResult := make(chan *Game, 1)
	timerDone := make(chan timerResult, 1)
	go func() {
		<-start
		requestResult <- manager.ModifiableGame(game.ID())
	}()
	go func() {
		<-start
		fired, err := manager.Internals().ForceNextTimerWithError()
		timerDone <- timerResult{fired: fired, err: err}
	}()
	close(start)

	select {
	case <-storage.entered:
	case <-time.After(time.Second):
		t.Fatal("cold checkout did not reach storage")
	}
	select {
	case <-storage.entered:
		close(storage.release)
		<-requestResult
		<-timerDone
		t.Fatal("timer and request started duplicate cold loads")
	case <-time.After(100 * time.Millisecond):
	}
	close(storage.release)

	requested := <-requestResult
	timer := <-timerDone
	if requested == nil {
		t.Fatal("request checkout returned nil")
	}
	if !timer.fired || timer.err != nil {
		t.Fatalf("timer completion = %t, %v; want successful fire", timer.fired, timer.err)
	}
	manager.modifiableGamesLock.RLock()
	resident := manager.modifiableGames[game.ID()]
	manager.modifiableGamesLock.RUnlock()
	if requested != resident {
		t.Fatal("request did not receive the timer's canonical mutable game")
	}
	if resident.Version() != version+1 {
		t.Fatalf("timer advanced canonical game to version %d; want %d", resident.Version(), version+1)
	}
}
