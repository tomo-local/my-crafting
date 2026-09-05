//go:build linux

package build_database

import (
	"os"
	"path"
	"syscall"
)

func createFileSync(file string) (*os.File, error) {
	fp, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}

	if err = syncDir(file); err != nil {
		_ = fp.Close()
		return nil, err
	}

	return fp, err
}

func syncDir(file string) error {
	flags := os.O_RDONLY | syscall.O_DIRECTORY
	// ディレクトリをファイル記述子として開く
	dirfd, err := syscall.Open(path.Dir(file), flags, 0o644)
	if err != nil {
		return err
	}
	defer syscall.Close(dirfd)
	// ディレクトリに対して fsync を実行
	return syscall.Fsync(dirfd)

}
