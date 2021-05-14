package library

import (
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestBatch(t *testing.T) {
	t.Run("MkDir", func(t *testing.T) {
		b := NewBatch()
		err := b.Do(MkDirAll("data/test"))
		if err != nil {
			t.Logf("Unexpected error: %s", err)
			return
		}
		b.Rollback()

		if _, err := os.Stat("data/test"); err == nil {
			t.Error("MkDir not reversed")
		}
	})
	t.Run("RollBack", func(t *testing.T) {
		b := NewBatch()
		var spyUndo bool
		err := b.Do(MkDirAll("data/test"))
		if err != nil {
			t.Logf("Unexpected error: %s", err)
			return
		}
		err = b.Do(NewAction("Allways wrong", func() error {
			return errors.New("I told you, it's an error")
		}).WithUndo(func() error {
			spyUndo = true
			return nil
		}))
		if err == nil {
			t.Errorf("Expecting error")
		}
		if spyUndo == false {
			t.Error("undo not called")
		}
		if _, err := os.Stat("data/test"); err == nil {
			t.Error("MkDir not reversed")
		}

	})

	t.Run("WriteFile", func(t *testing.T) {
		b := NewBatch()
		err := b.Do(MkDirAll("data/test"))
		if err != nil {
			t.Logf("Unexpected error: %s", err)
			return
		}

		want := []byte("hi mum!")

		err = b.Do(WriteFile("data/test/test.txt", want))
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}

		if got, err := os.ReadFile("data/test/test.txt"); err != nil {
			t.Errorf("Unexpected error: %s", err)
		} else {
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Expecting message to be %s, got %s", want, got)
			}
		}

		b.Rollback()
		if _, err := os.ReadFile("data/test/test.txt"); err == nil {
			t.Errorf("Expecting error, the file is still there!")
		}
		if _, err := os.Stat("data/test"); err == nil {
			t.Error("MkDir not reversed")
		}
	})

	t.Run("Wrong WriteFile", func(t *testing.T) {
		b := NewBatch()
		err := b.Do(MkDirAll("data/test"))
		if err != nil {
			t.Logf("Unexpected error: %s", err)
			return
		}

		want := []byte("hi mum!")

		err = b.Do(WriteFile("data/inexistant/test.txt", want))
		if err == nil {
			t.Errorf("Expecting an  error")
		}

		if _, err := os.ReadFile("data/inexistant/test.txt"); err == nil {
			t.Errorf("Expected error")
		}

		if _, err := os.ReadFile("data/test/test.txt"); err == nil {
			t.Errorf("Expecting error, the file is still there!")
		}
		if _, err := os.Stat("data/test"); err == nil {
			t.Error("MkDir not reversed")
		}
	})
}
