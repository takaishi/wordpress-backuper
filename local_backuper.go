package main

import (
	"io"
	"os"
	"path/filepath"
)

type LocalBackuper struct {
	source      string
	destination string
}

func (b *LocalBackuper) Run() error {
	err := b.copy(DB_DUMP_FILE)
	if err != nil {
		return err
	}

	err = b.copy(WORDPRESS_FILE)
	if err != nil {
		return err
	}

	return nil
}

func (b *LocalBackuper) copy(path string) error {
	src, err := os.Open(filepath.Join(b.source, path))
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(filepath.Join(b.destination, path))
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(src, dst)
	if err != nil {
		return err
	}

	return nil
}
