package memory

import (
	"context"
	"testing"

	"github.com/giantswarm/microstorage"
	"github.com/giantswarm/microstorage/storagetest"
)

func Test_List(t *testing.T) {
	config := DefaultConfig()
	newStorage, err := New(config)
	if err != nil {
		panic(err)
	}

	val := "my-val"

	err = newStorage.Create(context.TODO(), "key/one", val)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}
	err = newStorage.Create(context.TODO(), "key/two", val)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	values, err := newStorage.List(context.TODO(), "key")
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}
	if len(values) != 2 {
		t.Fatal("expected", 2, "got", len(values))
	}
}

func Test_List_Invalid(t *testing.T) {
	config := DefaultConfig()
	newStorage, err := New(config)
	if err != nil {
		panic(err)
	}

	val := "my-val"

	err = newStorage.Create(context.TODO(), "key/one", val)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}
	err = newStorage.Create(context.TODO(), "key/two", val)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	_, err = newStorage.List(context.TODO(), "ke")
	if !microstorage.IsNotFound(err) {
		t.Fatal("expected", true, "got", false)
	}
}

func Test_Storage(t *testing.T) {
	storage, err := New(DefaultConfig())
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}
	storagetest.Test(t, storage)
}
