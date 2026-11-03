package rest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer(t *testing.T, check func(r *http.Request)) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		check(r)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"strarr":["rest"]}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func initAuth(t *testing.T, config map[string]any) *authenticator {
	t.Helper()
	conf, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	a := &authenticator{}
	if err := a.Init(conf, "rest"); err != nil {
		t.Fatal(err)
	}
	return a
}

func TestCallEndpointSendsAuthTokenAndHeaders(t *testing.T) {
	var got http.Header
	srv := newTestServer(t, func(r *http.Request) { got = r.Header.Clone() })

	a := initAuth(t, map[string]any{
		"server_url":             srv.URL,
		"use_separate_endpoints": true,
		"auth_token":             "s3cret",
		"headers":                map[string]string{"X-Caller": "hedwig"},
	})
	if _, err := a.RestrictedTags(); err != nil {
		t.Fatal(err)
	}

	if v := got.Get("Authorization"); v != "Bearer s3cret" {
		t.Errorf("Authorization = %q, want %q", v, "Bearer s3cret")
	}
	if v := got.Get("X-Caller"); v != "hedwig" {
		t.Errorf("X-Caller = %q, want %q", v, "hedwig")
	}
	if v := got.Get("Content-Type"); v != "application/json" {
		t.Errorf("Content-Type = %q, want %q", v, "application/json")
	}
}

func TestCallEndpointWithoutAuthToken(t *testing.T) {
	var got http.Header
	srv := newTestServer(t, func(r *http.Request) { got = r.Header.Clone() })

	a := initAuth(t, map[string]any{"server_url": srv.URL})
	if _, err := a.RestrictedTags(); err != nil {
		t.Fatal(err)
	}

	if v := got.Get("Authorization"); v != "" {
		t.Errorf("Authorization = %q, want empty", v)
	}
}
