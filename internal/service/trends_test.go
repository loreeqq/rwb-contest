package service

import (
	"errors"
	"reflect"
	"testing"

	"rwb-contest/internal/config"
	"rwb-contest/internal/dto"
	"rwb-contest/internal/storage"
)

type fakeStorage struct {
	addedQuery     string
	topN           int
	topItems       []dto.TopItem
	stopWords      []string
	addStopWordArg string
	removeWordArg  string
	addStopWordErr error
}

func (f *fakeStorage) Add(query string) {
	f.addedQuery = query
}

func (f *fakeStorage) Top(n int) []dto.TopItem {
	f.topN = n
	return f.topItems
}

func (f *fakeStorage) AddStopWord(word string) error {
	f.addStopWordArg = word
	return f.addStopWordErr
}

func (f *fakeStorage) RemoveStopWord(word string) {
	f.removeWordArg = word
}

func (f *fakeStorage) ListStopWords() []string {
	return f.stopWords
}

func TestTrendsService_ProcessEvent(t *testing.T) {
	fs := &fakeStorage{}
	svc := NewTrendsService(fs)

	svc.ProcessEvent(dto.SearchEvent{Query: "iphone"})

	if fs.addedQuery != "iphone" {
		t.Fatalf("expected query iphone to be passed to storage, got %q", fs.addedQuery)
	}
}

func TestTrendsService_GetTop(t *testing.T) {
	expectedItems := []dto.TopItem{
		{Query: "iphone", Count: 3},
		{Query: "macbook", Count: 2},
	}

	fs := &fakeStorage{topItems: expectedItems}
	svc := NewTrendsService(fs)

	response := svc.GetTop(10)

	if fs.topN != 10 {
		t.Fatalf("expected top n 10 to be passed to storage, got %d", fs.topN)
	}

	if response.WindowSeconds != config.WindowSeconds {
		t.Fatalf("expected window_seconds %d, got %d", config.WindowSeconds, response.WindowSeconds)
	}

	if !reflect.DeepEqual(response.Items, expectedItems) {
		t.Fatalf("expected items %v, got %v", expectedItems, response.Items)
	}
}

func TestTrendsService_AddStopWord(t *testing.T) {
	fs := &fakeStorage{}
	svc := NewTrendsService(fs)

	if err := svc.AddStopWord("casino"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fs.addStopWordArg != "casino" {
		t.Fatalf("expected casino to be passed to storage, got %q", fs.addStopWordArg)
	}
}

func TestTrendsService_AddStopWordReturnsStorageError(t *testing.T) {
	expectedErr := storage.ErrWordAlreadyExist
	fs := &fakeStorage{addStopWordErr: expectedErr}
	svc := NewTrendsService(fs)

	err := svc.AddStopWord("casino")

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestTrendsService_RemoveStopWord(t *testing.T) {
	fs := &fakeStorage{}
	svc := NewTrendsService(fs)

	svc.RemoveStopWord("casino")

	if fs.removeWordArg != "casino" {
		t.Fatalf("expected casino to be passed to storage, got %q", fs.removeWordArg)
	}
}

func TestTrendsService_ListStopWords(t *testing.T) {
	expectedWords := []string{"casino", "spam"}
	fs := &fakeStorage{stopWords: expectedWords}
	svc := NewTrendsService(fs)

	words := svc.ListStopWords()

	if !reflect.DeepEqual(words, expectedWords) {
		t.Fatalf("expected stop-words %v, got %v", expectedWords, words)
	}
}
