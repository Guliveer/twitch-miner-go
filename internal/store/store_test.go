package store_test

import (
	"testing"

	"github.com/Guliveer/twitch-miner-go/internal/store"
)

// Compile-time check: NoopStore must satisfy Store.
var _ store.Store = store.NoopStore{}

func TestNoopStore_ListAccountsReturnsEmpty(t *testing.T) {
	s := store.NoopStore{}
	rows, err := s.ListAccounts()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Errorf("expected empty slice, got %d rows", len(rows))
	}
}

func TestNoopStore_GetAccountReturnsNil(t *testing.T) {
	s := store.NoopStore{}
	row, err := s.GetAccount("anyone")
	if err != nil {
		t.Fatal(err)
	}
	if row != nil {
		t.Errorf("expected nil, got %+v", row)
	}
}

func TestNoopStore_UpsertIsNoop(t *testing.T) {
	s := store.NoopStore{}
	if err := s.UpsertAccount(store.AccountRow{Username: "x"}); err != nil {
		t.Fatal(err)
	}
}

func TestNoopStore_DeleteIsNoop(t *testing.T) {
	s := store.NoopStore{}
	if err := s.DeleteAccount("x"); err != nil {
		t.Fatal(err)
	}
}

func TestNoopStore_ChangesIsNil(t *testing.T) {
	s := store.NoopStore{}
	if s.Changes() != nil {
		t.Error("expected nil channel from NoopStore.Changes()")
	}
}
