package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/Sothatsit/work-stream/internal/api"
)

const testBusyTimeout = time.Second

func openStoreAt(
	t *testing.T,
	path string,
	busyTimeout time.Duration,
) *Store {
	t.Helper()
	store, err := Open(context.Background(), path, busyTimeout)
	if err != nil {
		t.Fatalf("opening store: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("closing store: %v", err)
		}
	})
	return store
}

func openTestStore(t *testing.T) *Store {
	t.Helper()
	return openStoreAt(
		t, filepath.Join(t.TempDir(), "test.db"), testBusyTimeout,
	)
}

func mustAdd(
	t *testing.T,
	store *Store,
	typ string,
	subject string,
	body string,
	metadata map[string]string,
) api.Entry {
	t.Helper()
	entry, err := store.Add(context.Background(), api.AddEntryRequest{
		Type: typ, Subject: subject, Body: body, Metadata: metadata,
	})
	if err != nil {
		t.Fatalf("adding entry: %v", err)
	}
	return entry
}

func TestGeneratedColumnsFollowWrites(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)
	entry := mustAdd(
		t, store, "TODO", "COUNT Æ", "BODY Σ",
		map[string]string{"jira": "QUACK Æ"},
	)
	assertFolded := func(
		wantType string,
		wantSubject string,
		wantBody string,
		wantValue string,
	) {
		t.Helper()
		var typ, subject, body, value string
		err := store.db.QueryRowContext(ctx, `
			SELECT e.type_lower, e.subject_lower, e.body_lower,
				metadata.value_lower
			FROM entries AS e
			JOIN entry_metadata AS metadata ON metadata.entry_id = e.id
			WHERE e.id = ?`, entry.ID,
		).Scan(&typ, &subject, &body, &value)
		if err != nil {
			t.Fatalf("reading folded values: %v", err)
		}
		if typ != wantType || subject != wantSubject ||
			body != wantBody || value != wantValue {
			t.Fatalf(
				"folded values = %q, %q, %q, %q; want %q, %q, %q, %q",
				typ, subject, body, value,
				wantType, wantSubject, wantBody, wantValue,
			)
		}
	}
	assertFolded("todo", "count Æ", "body Σ", "quack Æ")

	entryType := "NOTE"
	subject := "CHANGED Æ"
	body := "CHANGED Σ"
	if _, err := store.Edit(ctx, entry.ID, api.EditEntryRequest{
		Type: &entryType, Subject: &subject, Body: &body,
		Metadata: map[string]string{"jira": "CHANGED Æ"},
	}); err != nil {
		t.Fatalf("editing entry: %v", err)
	}
	assertFolded("note", "changed Æ", "changed Σ", "changed Æ")
}

func TestMutationRollsBackWhenResultCannotBeRead(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)
	_, err := store.db.ExecContext(ctx, `
		CREATE TRIGGER corrupt_new_entry AFTER INSERT ON entries
		BEGIN
			UPDATE entries SET created_utc = 'bad' WHERE id = NEW.id;
		END`)
	if err != nil {
		t.Fatalf("creating trigger: %v", err)
	}
	_, err = store.Add(ctx, api.AddEntryRequest{
		Type: "note", Subject: "must roll back",
	})
	if err == nil {
		t.Fatal("add succeeded despite unreadable result")
	}
	var count int
	if err := store.db.QueryRowContext(
		ctx, `SELECT count(*) FROM entries`,
	).Scan(&count); err != nil {
		t.Fatalf("counting entries: %v", err)
	}
	if count != 0 {
		t.Fatalf("failed add committed %d entries", count)
	}
}

func TestDeletedIDsAreNotReused(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)
	first := mustAdd(t, store, "note", "first", "", nil)
	if err := store.Delete(ctx, first.ID); err != nil {
		t.Fatalf("deleting entry: %v", err)
	}
	second := mustAdd(t, store, "note", "second", "", nil)
	if second.ID <= first.ID {
		t.Fatalf("id %d reused after deleting %d", second.ID, first.ID)
	}
}

func TestDeleteCascadesMetadata(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)
	entry := mustAdd(t, store, "note", "entry", "",
		map[string]string{"jira": "QUACK-1"})
	if err := store.Delete(ctx, entry.ID); err != nil {
		t.Fatalf("deleting entry: %v", err)
	}
	var count int
	if err := store.db.QueryRowContext(ctx, `
		SELECT count(*) FROM entry_metadata`,
	).Scan(&count); err != nil {
		t.Fatalf("counting metadata: %v", err)
	}
	if count != 0 {
		t.Fatalf("metadata count = %d, want 0", count)
	}
}

func TestMetadataMutationsReturnTouchedEntry(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name string
		set  map[string]string
		want map[string]string
	}{
		{
			"add",
			map[string]string{"review": "done"},
			map[string]string{"jira": "QUACK-1", "review": "done"},
		},
		{
			"overwrite",
			map[string]string{"jira": "QUACK-2"},
			map[string]string{"jira": "QUACK-2"},
		},
		{
			"remove",
			map[string]string{"jira": ""},
			map[string]string{},
		},
		{
			"remove a key that was never set",
			map[string]string{"review": ""},
			map[string]string{"jira": "QUACK-1"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := openTestStore(t)
			entry := mustAdd(t, store, "note", "entry", "",
				map[string]string{"jira": "QUACK-1"})
			oldTime := "2000-01-01T00:00:00.000000000Z"
			if _, err := store.db.ExecContext(ctx, `
				UPDATE entries SET modified_utc = ? WHERE id = ?`,
				oldTime, entry.ID,
			); err != nil {
				t.Fatalf("setting old modification time: %v", err)
			}
			got, err := store.Edit(ctx, entry.ID, api.EditEntryRequest{
				Metadata: test.set,
			})
			if err != nil {
				t.Fatalf("changing metadata: %v", err)
			}
			oldModified, err := time.Parse(time.RFC3339, oldTime)
			if err != nil {
				t.Fatal(err)
			}
			if !got.Modified.After(oldModified) {
				t.Fatalf("modified was not refreshed: %v", got.Modified)
			}
			if len(got.Metadata) != len(test.want) {
				t.Fatalf("metadata = %v, want %v", got.Metadata, test.want)
			}
			for key, value := range test.want {
				if got.Metadata[key] != value {
					t.Fatalf("metadata = %v, want %v", got.Metadata, test.want)
				}
			}
		})
	}
}

func TestCanceledMutationDoesNotWrite(t *testing.T) {
	store := openTestStore(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := store.Add(ctx, api.AddEntryRequest{
		Type: "note", Subject: "canceled",
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("add error = %v, want context.Canceled", err)
	}
	var count int
	if err := store.db.QueryRow(`SELECT count(*) FROM entries`).Scan(
		&count,
	); err != nil {
		t.Fatalf("counting entries: %v", err)
	}
	if count != 0 {
		t.Fatalf("canceled add wrote %d entries", count)
	}
}

func TestBusyDatabaseIsRecognized(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "busy.db")
	first := openStoreAt(t, path, testBusyTimeout)
	second := openStoreAt(t, path, 5*time.Millisecond)
	entry := mustAdd(t, first, "note", "locked", "", nil)

	tx, err := first.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("starting lock transaction: %v", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
		UPDATE entries SET subject = subject WHERE id = ?`, entry.ID,
	); err != nil {
		t.Fatalf("locking database: %v", err)
	}
	_, err = second.Add(ctx, api.AddEntryRequest{
		Type: "note", Subject: "blocked",
	})
	if !IsBusy(err) {
		t.Fatalf("add error = %v, want SQLite busy/locked", err)
	}
	if !IsBusy(errors.Join(errors.New("wrapper"), err)) {
		t.Fatal("IsBusy did not inspect wrapped error")
	}
}

func TestConcurrentMetadataWritesSerialize(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)
	first := mustAdd(t, store, "note", "first", "", nil)
	second := mustAdd(t, store, "note", "second", "", nil)

	start := make(chan struct{})
	results := make(chan error, 2)
	for _, id := range []int64{first.ID, second.ID} {
		go func() {
			<-start
			_, err := store.Edit(ctx, id, api.EditEntryRequest{
				Metadata: map[string]string{"review": "done"},
			})
			results <- err
		}()
	}
	close(start)
	for range 2 {
		if err := <-results; err != nil {
			t.Fatalf("parallel metadata write: %v", err)
		}
	}

	start = make(chan struct{})
	for _, value := range []string{"QUACK-1", "QUACK-2"} {
		go func() {
			<-start
			_, err := store.Edit(ctx, first.ID, api.EditEntryRequest{
				Metadata: map[string]string{"jira": value},
			})
			results <- err
		}()
	}
	close(start)
	for range 2 {
		if err := <-results; err != nil {
			t.Fatalf("competing metadata write: %v", err)
		}
	}
	got, err := store.Get(ctx, first.ID)
	if err != nil {
		t.Fatalf("getting entry: %v", err)
	}
	if got.Metadata["jira"] != "QUACK-1" &&
		got.Metadata["jira"] != "QUACK-2" {
		t.Fatalf("jira = %q, want one of the competing values",
			got.Metadata["jira"])
	}
}

func TestMetadataLimitStopsEntryGrowth(t *testing.T) {
	store := openTestStore(t)
	metadata := make(map[string]string, api.MaxMetadata)
	for index := range api.MaxMetadata {
		metadata[fmt.Sprintf("key-%d", index)] = "value"
	}
	entry := mustAdd(t, store, "note", "bounded", "", metadata)
	ctx := context.Background()
	_, err := store.Edit(ctx, entry.ID, api.EditEntryRequest{
		Metadata: map[string]string{"overflow": "value"},
	})
	if !errors.Is(err, ErrMetadataLimit) {
		t.Fatalf("Edit error = %v, want ErrMetadataLimit", err)
	}

	edited, err := store.Edit(ctx, entry.ID, api.EditEntryRequest{
		Metadata: map[string]string{"key-0": "replaced"},
	})
	if err != nil {
		t.Fatalf("overwriting a metadata value at the limit: %v", err)
	}
	if got := edited.Metadata["key-0"]; got != "replaced" {
		t.Fatalf("key-0 = %q, want replaced", got)
	}
}

var _ databaseReader = (*sql.DB)(nil)
