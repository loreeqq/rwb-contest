package storage

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"rwb-contest/internal/config"
)

func TestRingBufferStorage_AddAndTop(t *testing.T) {
	s := NewRingBufferStorage()

	s.Add("iphone")
	s.Add("iphone")
	s.Add("macbook")

	top := s.Top(10)

	if len(top) != 2 {
		t.Fatalf("expected 2 items, got %d", len(top))
	}

	if top[0].Query != "iphone" {
		t.Fatalf("expected iphone to be first, got %q", top[0].Query)
	}

	if top[0].Count != 2 {
		t.Fatalf("expected iphone count 2, got %d", top[0].Count)
	}

	if top[1].Query != "macbook" {
		t.Fatalf("expected macbook to be second, got %q", top[1].Query)
	}

	if top[1].Count != 1 {
		t.Fatalf("expected macbook count 1, got %d", top[1].Count)
	}
}

func TestRingBufferStorage_NormalizeQuery(t *testing.T) {
	s := NewRingBufferStorage()

	s.Add("  IPHONE  ")
	s.Add("iphone")

	top := s.Top(10)

	if len(top) != 1 {
		t.Fatalf("expected 1 item after normalization, got %d", len(top))
	}

	if top[0].Query != "iphone" {
		t.Fatalf("expected normalized query iphone, got %q", top[0].Query)
	}

	if top[0].Count != 2 {
		t.Fatalf("expected count 2, got %d", top[0].Count)
	}
}

func TestRingBufferStorage_IgnoreEmptyQuery(t *testing.T) {
	s := NewRingBufferStorage()

	s.Add("")
	s.Add("   ")

	top := s.Top(10)

	if len(top) != 0 {
		t.Fatalf("expected empty top, got %d items", len(top))
	}
}

func TestRingBufferStorage_TopLimit(t *testing.T) {
	s := NewRingBufferStorage()

	s.Add("iphone")
	s.Add("macbook")
	s.Add("airpods")

	top := s.Top(2)

	if len(top) != 2 {
		t.Fatalf("expected 2 items, got %d", len(top))
	}
}

func TestRingBufferStorage_TopDefaultLimitForInvalidN(t *testing.T) {
	s := NewRingBufferStorage()

	for i := 0; i < 15; i++ {
		s.Add(fmt.Sprintf("query-%d", i))
	}

	top := s.Top(0)

	if len(top) != 10 {
		t.Fatalf("expected default limit 10, got %d", len(top))
	}
}

func TestRingBufferStorage_StopWordFiltersQuery(t *testing.T) {
	s := NewRingBufferStorage()

	if err := s.AddStopWord("casino"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s.Add("best casino")
	s.Add("iphone")

	top := s.Top(10)

	if len(top) != 1 {
		t.Fatalf("expected 1 item, got %d", len(top))
	}

	if top[0].Query != "iphone" {
		t.Fatalf("expected only iphone in top, got %q", top[0].Query)
	}
}

func TestRingBufferStorage_StopWordsAreNormalizedAndSorted(t *testing.T) {
	s := NewRingBufferStorage()

	if err := s.AddStopWord("  Zoo  "); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := s.AddStopWord("Apple"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	words := s.ListStopWords()

	if len(words) != 2 {
		t.Fatalf("expected 2 stop-words, got %d", len(words))
	}

	if words[0] != "apple" || words[1] != "zoo" {
		t.Fatalf("expected sorted normalized stop-words [apple zoo], got %v", words)
	}
}

func TestRingBufferStorage_DuplicateStopWordReturnsError(t *testing.T) {
	s := NewRingBufferStorage()

	if err := s.AddStopWord("casino"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err := s.AddStopWord(" casino ")

	if !errors.Is(err, ErrWordAlreadyExist) {
		t.Fatalf("expected ErrWordAlreadyExist, got %v", err)
	}
}

func TestRingBufferStorage_RemoveStopWord(t *testing.T) {
	s := NewRingBufferStorage()

	if err := s.AddStopWord("casino"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s.RemoveStopWord("casino")
	s.Add("casino")

	top := s.Top(10)

	if len(top) != 1 {
		t.Fatalf("expected query to be counted after stop-word removal, got %d items", len(top))
	}

	if top[0].Query != "casino" {
		t.Fatalf("expected casino, got %q", top[0].Query)
	}
}

func TestRingBufferStorage_AntiSpamLimit(t *testing.T) {
	s := NewRingBufferStorage()

	for i := 0; i < config.MaxQueryPerSecond+25; i++ {
		s.Add("iphone")
	}

	top := s.Top(10)

	if len(top) != 1 {
		t.Fatalf("expected 1 item, got %d", len(top))
	}

	if top[0].Count != config.MaxQueryPerSecond {
		t.Fatalf("expected count to be capped at %d, got %d", config.MaxQueryPerSecond, top[0].Count)
	}
}

func TestRingBufferStorage_ConcurrentAccess(t *testing.T) {
	s := NewRingBufferStorage()

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Add("iphone")
		}()
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.Top(10)
		}()
	}

	wg.Wait()

	top := s.Top(10)

	if len(top) != 1 {
		t.Fatalf("expected 1 item, got %d", len(top))
	}

	if top[0].Count <= 0 {
		t.Fatalf("expected positive count, got %d", top[0].Count)
	}
}
