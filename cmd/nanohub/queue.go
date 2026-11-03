package main

// QA Wolf patch (TITAN-5741). Same path and status codes as the queue API
// proposed upstream in micromdm/nanomdm#269. Drop it at the NanoHUB upgrade
// that ships that API.

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/micromdm/nanolib/log"
	"github.com/micromdm/nanolib/log/ctxlog"
	"github.com/micromdm/nanomdm/api"
	"github.com/micromdm/nanomdm/mdm"
	mdmstorage "github.com/micromdm/nanomdm/storage"
)

const apiEndpointQueue = "/queue/" // note trailing slash

type queueAPIResult struct {
	Status map[string]*api.Error `json:"status,omitempty"`
}

// queueClearHandler clears the command queues of the comma-separated device
// IDs in the URL path, the same way NanoMDM clears them at a new enrollment.
// It returns 204 when every clear works, 207 when some fail and 500 when all
// fail. Strip the endpoint prefix before it.
func queueClearHandler(store mdmstorage.CommandAndReportResultsStore, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.Header().Set("Allow", http.MethodDelete)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		logger := ctxlog.Logger(r.Context(), logger)

		var ids []string
		for _, id := range strings.Split(r.URL.Path, ",") {
			if id = strings.TrimSpace(id); id != "" {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			writeQueueJSON(w, http.StatusBadRequest, map[string]string{"error": "missing enrollment id(s)"})
			return
		}

		result := queueAPIResult{Status: make(map[string]*api.Error)}
		var failed int
		for _, id := range ids {
			mr := mdm.NewRequestWithContext(r.Context(), nil)
			mr.EnrollID = &mdm.EnrollID{Type: mdm.Device, ID: id}
			if err := store.ClearQueue(mr); err != nil {
				logger.Info("msg", "clearing queue", "id", id, "err", err)
				result.Status[id] = api.NewError(err)
				failed++
				continue
			}
			logger.Info("msg", "cleared queue", "id", id)
		}

		switch failed {
		case 0:
			w.WriteHeader(http.StatusNoContent)
		case len(ids):
			writeQueueJSON(w, http.StatusInternalServerError, result)
		default:
			writeQueueJSON(w, http.StatusMultiStatus, result)
		}
	}
}

func writeQueueJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
