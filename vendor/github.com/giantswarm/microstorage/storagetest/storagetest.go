package storagetest

import (
	"context"
	"testing"

	"github.com/giantswarm/microstorage"
)

// Test is Storage conformance test.
func Test(t *testing.T, storage microstorage.Storage) {
	key := "test-key"
	value := "test-value"

	ok, err := storage.Exists(context.TODO(), key)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}
	if ok {
		t.Fatal("expected", false, "got", true)
	}

	err = storage.Put(context.TODO(), key, value)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	// Second Set call should pass.
	err = storage.Put(context.TODO(), key, value)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	ok, err = storage.Exists(context.TODO(), key)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}
	if !ok {
		t.Fatal("expected", true, "got", false)
	}

	v, err := storage.Search(context.TODO(), key)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}
	if v != value {
		t.Fatal("expected", value, "got", v)
	}

	err = storage.Delete(context.TODO(), key)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	ok, err = storage.Exists(context.TODO(), key)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}
	if ok {
		t.Fatal("expected", false, "got", true)
	}

	v, err = storage.Search(context.TODO(), key)
	if !microstorage.IsNotFound(err) {
		t.Fatal("expected", true, "got", false)
	}
	if v != "" {
		t.Fatal("expected", "empty string", "got", v)
	}
}
