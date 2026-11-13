package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStopPollingPersistsAndCancelsCycle(t *testing.T) {
	setupTestDB(t)
	if _, err := SaveNetworkPollingConfig(SaveNetworkPollingConfigRequest{Enabled: true, IntervalSeconds: 60}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pollingMu.Lock()
	pollingCancel = cancel
	pollingMu.Unlock()
	t.Cleanup(func() {
		pollingMu.Lock()
		pollingCancel = nil
		pollingMu.Unlock()
		select {
		case <-pollingWake:
		default:
		}
	})
	recorder := httptest.NewRecorder()
	stopNetworkPolling(recorder, httptest.NewRequest(http.MethodPost, "/api/network-status/stop", nil))
	if recorder.Code != 200 {
		t.Fatalf("stop: %d %s", recorder.Code, recorder.Body.String())
	}
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Fatal("active polling context not canceled")
	}
	config, err := GetNetworkPollingConfig()
	if err != nil || config.Enabled {
		t.Fatalf("polling still enabled: %+v %v", config, err)
	}
	select {
	case <-pollingWake:
	default:
		t.Fatal("poller was not notified")
	}
}

func TestCanceledPollDoesNotStartScan(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := scanTeamOpenPorts(ctx, "unused"); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled scan returned %v", err)
	}
}
