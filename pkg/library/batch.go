package library

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type runFn func() error

type Action struct {
	name string
	do   runFn
	undo runFn
}

var nop = &Action{
	name: "Nop",
}

func NewAction(name string, do runFn) *Action {
	return &Action{
		name: name,
		do:   do,
	}
}

func (a *Action) WithUndo(undo runFn) *Action {
	a.undo = undo
	return a
}

type Batch struct {
	l   Logger
	Log []*Action
}

func NewBatch() *Batch {
	return &Batch{
		l: nullLogger{},
	}
}
func (b *Batch) WithLogger(l Logger) *Batch {
	b.l = l
	return b
}

func (b *Batch) Do(a *Action) error {
	if a.do == nil {
		return nil
	}
	b.Log = append(b.Log, a)
	err := a.do()
	if err != nil {
		b.l.Printf("[BATCH] %s: %s, rollback batch of actions", a.name, err)
		b.Rollback()
	}
	b.l.Printf("[BATCH] %s done", a.name)
	return err
}

func (b *Batch) Rollback() error {
	for len(b.Log) > 0 {
		a := b.Log[len(b.Log)-1]
		if a.undo != nil {
			a.undo()
			b.l.Printf("[BATCH] %s undone", a.name)
		}
		b.Log = b.Log[:len(b.Log)-1]
	}
	return nil
}

func MkDir(d string) *Action {
	if _, err := os.Stat(d); err == nil {
		return nop
	}
	return NewAction(fmt.Sprintf("MkDir %q", d), func() error {
		return os.Mkdir(d, 0777)
	}).WithUndo(func() error {
		return os.Remove(d)
	})
}

func MkDirAll(alldir string) *Action {
	batch := NewBatch()

	return NewAction(fmt.Sprintf("MkDirAll %q", alldir),
		func() error {
			var err error
			alldir, err = filepath.Abs(alldir)
			if err != nil {
				return err
			}
			path := filepath.VolumeName(alldir) + string(os.PathSeparator)
			alldir = strings.TrimPrefix(alldir, path)
			for _, d := range strings.Split(alldir, string(os.PathSeparator)) {
				if d == "" {
					continue
				}
				path = filepath.Join(path, d)
				if _, err = os.Stat(path); err == nil {
					continue
				}
				err = batch.Do(MkDir(path))
				if err != nil {
					return err
				}
			}
			return nil
		}).WithUndo(func() error {
		return batch.Rollback()
	})
}

func WriteFile(FileName string, content []byte) *Action {
	return NewAction(
		fmt.Sprintf("WriteFile %q", FileName),
		func() error {
			f, err := os.Create(FileName)
			if err != nil {
				return err
			}
			_, err = f.Write(content)
			if err != nil {
				return err
			}
			return f.Close()
		}).WithUndo(
		func() error {
			return os.Remove(FileName)
		})
}
