package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/micromdm/nanolib/log"
	"github.com/micromdm/nanomdm/mdm"
	mdmstorage "github.com/micromdm/nanomdm/storage"
	mdminmem "github.com/micromdm/nanomdm/storage/inmem"
)

func newQueueTestHandler(store mdmstorage.CommandAndReportResultsStore) http.Handler {
	mux := http.NewServeMux()
	mux.Handle(apiEndpointQueue, http.StripPrefix(apiEndpointQueue, queueClearHandler(store, log.NopLogger)))
	return mux
}

func doQueueRequest(t *testing.T, h http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func nextCommand(t *testing.T, store mdmstorage.CommandAndReportResultsStore, id string) *mdm.Command {
	t.Helper()
	r := mdm.NewRequestWithContext(context.Background(), nil)
	r.EnrollID = &mdm.EnrollID{Type: mdm.Device, ID: id}
	cmd, err := store.RetrieveNextCommand(r, false)
	if err != nil {
		t.Fatal(err)
	}
	return cmd
}

func TestQueueClear(t *testing.T) {
	ctx := context.Background()
	store := mdminmem.New()
	h := newQueueTestHandler(store)

	for _, c := range []struct{ uuid, id string }{{"CMD1", "DEV1"}, {"CMD2", "DEV2"}, {"CMD3", "DEV3"}} {
		cmd := &mdm.Command{CommandUUID: c.uuid, Raw: []byte("raw")}
		cmd.Command.RequestType = "InstallApplication"
		if _, err := store.EnqueueCommand(ctx, []string{c.id}, cmd); err != nil {
			t.Fatal(err)
		}
	}
	if cmd := nextCommand(t, store, "DEV1"); cmd == nil || cmd.CommandUUID != "CMD1" {
		t.Fatalf("before clear: got %v, want CMD1", cmd)
	}

	if rec := doQueueRequest(t, h, http.MethodDelete, "/queue/DEV1"); rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE one: got %d, want 204", rec.Code)
	}
	if cmd := nextCommand(t, store, "DEV1"); cmd != nil {
		t.Errorf("DEV1 after clear: got %s, want empty queue", cmd.CommandUUID)
	}
	if cmd := nextCommand(t, store, "DEV2"); cmd == nil {
		t.Error("DEV2 lost its command; only DEV1 was cleared")
	}

	if rec := doQueueRequest(t, h, http.MethodDelete, "/queue/DEV2,DEV3"); rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE list: got %d, want 204", rec.Code)
	}
	for _, id := range []string{"DEV2", "DEV3"} {
		if cmd := nextCommand(t, store, id); cmd != nil {
			t.Errorf("%s after clear: got %s, want empty queue", id, cmd.CommandUUID)
		}
	}

	// An unknown or already empty queue is not an error.
	if rec := doQueueRequest(t, h, http.MethodDelete, "/queue/DEV1"); rec.Code != http.StatusNoContent {
		t.Errorf("DELETE empty queue: got %d, want 204", rec.Code)
	}
}

func TestQueueClearRequests(t *testing.T) {
	h := newQueueTestHandler(mdminmem.New())

	if rec := doQueueRequest(t, h, http.MethodGet, "/queue/DEV1"); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET: got %d, want 405", rec.Code)
	}
	if rec := doQueueRequest(t, h, http.MethodDelete, "/queue/"); rec.Code != http.StatusBadRequest {
		t.Errorf("DELETE without id: got %d, want 400", rec.Code)
	}
}

type failingQueueStore struct {
	mdmstorage.CommandAndReportResultsStore
	fail map[string]bool
}

func (s *failingQueueStore) ClearQueue(r *mdm.Request) error {
	if s.fail[r.ID] {
		return errors.New("db down")
	}
	return nil
}

func TestQueueClearErrors(t *testing.T) {
	h := newQueueTestHandler(&failingQueueStore{fail: map[string]bool{"BAD1": true, "BAD2": true}})

	for _, tc := range []struct {
		path     string
		wantCode int
		wantBad  []string
	}{
		{"/queue/BAD1,BAD2", http.StatusInternalServerError, []string{"BAD1", "BAD2"}},
		{"/queue/OK1,BAD1", http.StatusMultiStatus, []string{"BAD1"}},
	} {
		rec := doQueueRequest(t, h, http.MethodDelete, tc.path)
		if rec.Code != tc.wantCode {
			t.Errorf("%s: got %d, want %d", tc.path, rec.Code, tc.wantCode)
			continue
		}
		var res struct {
			Status map[string]string `json:"status"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("%s: decoding body: %v", tc.path, err)
		}
		if len(res.Status) != len(tc.wantBad) {
			t.Errorf("%s: status %v, want errors for %v", tc.path, res.Status, tc.wantBad)
		}
		for _, id := range tc.wantBad {
			if res.Status[id] == "" {
				t.Errorf("%s: no error for %s in %v", tc.path, id, res.Status)
			}
		}
	}
}
