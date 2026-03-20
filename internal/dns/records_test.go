package dns

import (
	"testing"
)

func newStoreWithRecords(records map[string]string) *RecordStore {
	s := &RecordStore{
		byFile: make(map[string]RecordFile),
		merged: make(map[string]string),
		empty:  make(chan struct{}),
	}
	for k, v := range records {
		s.merged[k] = v
	}
	return s
}

func TestLookup_ExactMatch(t *testing.T) {
	s := newStoreWithRecords(map[string]string{
		"api.sew.local": "10.0.0.1",
	})
	ip, ok := s.Lookup("api.sew.local")
	if !ok || ip != "10.0.0.1" {
		t.Fatalf("expected 10.0.0.1, got %q (ok=%v)", ip, ok)
	}
}

func TestLookup_ExactMatchCaseInsensitive(t *testing.T) {
	s := newStoreWithRecords(map[string]string{
		"api.sew.local": "10.0.0.1",
	})
	ip, ok := s.Lookup("API.SEW.LOCAL")
	if !ok || ip != "10.0.0.1" {
		t.Fatalf("expected 10.0.0.1, got %q (ok=%v)", ip, ok)
	}
}

func TestLookup_WildcardMatchSingleLabel(t *testing.T) {
	s := newStoreWithRecords(map[string]string{
		"*.kafka.sew.local": "10.0.0.2",
	})
	ip, ok := s.Lookup("demo.kafka.sew.local")
	if !ok || ip != "10.0.0.2" {
		t.Fatalf("expected 10.0.0.2, got %q (ok=%v)", ip, ok)
	}
}

func TestLookup_WildcardMatchHyphenatedLabel(t *testing.T) {
	s := newStoreWithRecords(map[string]string{
		"*.kafka.sew.local": "10.0.0.2",
	})
	ip, ok := s.Lookup("broker-0-demo.kafka.sew.local")
	if !ok || ip != "10.0.0.2" {
		t.Fatalf("expected 10.0.0.2, got %q (ok=%v)", ip, ok)
	}
}

func TestLookup_ExactTakesPriorityOverWildcard(t *testing.T) {
	s := newStoreWithRecords(map[string]string{
		"*.kafka.sew.local":   "10.0.0.2",
		"demo.kafka.sew.local": "10.0.0.3",
	})
	ip, ok := s.Lookup("demo.kafka.sew.local")
	if !ok || ip != "10.0.0.3" {
		t.Fatalf("expected exact match 10.0.0.3, got %q (ok=%v)", ip, ok)
	}
}

func TestLookup_NoMatch(t *testing.T) {
	s := newStoreWithRecords(map[string]string{
		"*.kafka.sew.local": "10.0.0.2",
	})
	_, ok := s.Lookup("api.sew.local")
	if ok {
		t.Fatal("expected no match for api.sew.local against *.kafka.sew.local")
	}
}

func TestLookup_NoMatchBareHostname(t *testing.T) {
	s := newStoreWithRecords(map[string]string{
		"*.kafka.sew.local": "10.0.0.2",
	})
	_, ok := s.Lookup("kafka.sew.local")
	if ok {
		t.Fatal("expected no match for kafka.sew.local (wildcard requires a label before the pattern)")
	}
}

func TestLookup_EmptyStore(t *testing.T) {
	s := newStoreWithRecords(map[string]string{})
	_, ok := s.Lookup("anything.sew.local")
	if ok {
		t.Fatal("expected no match in empty store")
	}
}
