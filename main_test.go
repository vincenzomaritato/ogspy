package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ----------------------------------------------------------------------------
// Unit tests for helper functions
// ----------------------------------------------------------------------------

func TestParseOG(t *testing.T) {
	html := `
	<!doctype html>
	<html><head>
	<meta property="og:title" content="Hello">
	<meta property="og:type"  content="article">
	<meta property="og:image" content="https://cdn.example.com/img.jpg">
	<meta property="og:url"   content="https://example.com">
	</head><body></body></html>`
	got := parseOG(html)

	want := map[string]string{
		"title": "Hello",
		"type":  "article",
		"image": "https://cdn.example.com/img.jpg",
		"url":   "https://example.com",
	}
	if len(got) != len(want) {
		t.Fatalf("parseOG len = %d, want %d", len(got), len(want))
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("parseOG[%q] = %q, want %q", k, got[k], v)
		}
	}
}

func TestDiffMaps(t *testing.T) {
	old := map[string]string{"title": "Old", "image": "A"}
	new := map[string]string{"title": "New", "image": "A", "type": "article"}
	diff := diffMaps(old, new)

	// title should differ, image identical, type added
	if _, ok := diff["title"]; !ok {
		t.Error("diffMaps: expected difference on title")
	}
	if _, ok := diff["image"]; ok {
		t.Error("diffMaps: unexpected diff on identical image")
	}
	if _, ok := diff["type"]; !ok {
		t.Error("diffMaps: expected diff on added type key")
	}
}

func TestSemanticValidate(t *testing.T) {
	og := map[string]string{
		"type":  "article",
		"image": "http://insecure/img.jpg",
	}
	warns := semanticValidate(og)
	if len(warns) == 0 {
		t.Fatal("semanticValidate: expected warnings, got none")
	}
}

// ----------------------------------------------------------------------------
// Integration test for fetchHTML using httptest server
// ----------------------------------------------------------------------------

func TestFetchHTML(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><head><title>x</title></head><body>ok</body></html>"))
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	doc, err := fetchHTML(ctx, srv.URL)
	if err != nil {
		t.Fatalf("fetchHTML error: %v", err)
	}
	if !strings.Contains(doc, "<title>x</title>") {
		t.Errorf("fetchHTML output mismatch")
	}
}
