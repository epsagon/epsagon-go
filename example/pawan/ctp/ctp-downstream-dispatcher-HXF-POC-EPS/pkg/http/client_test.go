package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Get__Success(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer ts.Close()

	client := NewClient(
		WithHTTPTimeout(defaultHTTPTimeout),
	)
	body, err := client.Get(ts.URL, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	expected := []byte("OK")
	if !bytes.Equal(body, expected) {
		t.Fatalf("bad: %v", body)
	}
}

func TestClient_Get__Error(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	}))
	defer ts.Close()

	client := NewClient(
		WithHTTPTimeout(defaultHTTPTimeout),
	)
	body, err := client.Get(ts.URL, nil)
	if err == nil {
		t.Fatalf("err: %v", err)
	}

	expected := []byte("Not found")
	if !bytes.Equal(body, expected) {
		t.Fatalf("bad: %v", body)
	}
}

func TestClient_Post(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer ts.Close()

	client := NewClient(
		WithHTTPTimeout(defaultHTTPTimeout),
	)

	customer := readJSONFile("../../test/resources/customer/non_servco.json")

	body, err := client.Post(ts.URL, bytes.NewBuffer(customer), nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	expected := []byte("OK")
	if !bytes.Equal(body, expected) {
		t.Fatalf("bad: %v", body)
	}
}

func readJSONFile(filename string) []byte {
	bts, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Print(err)
	}
	return bts
}
