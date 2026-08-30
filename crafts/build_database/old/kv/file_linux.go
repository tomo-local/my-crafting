package main

import (
	"os"
	"syscall"
	"path"
)

func createFileSync(file string) (*os.File, error){
	// ファイルを開く（または作成する）
	fp, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
			return nil, err
	}
	// 親ディレクトリを同期する
	if err = syncDir(path.Base(file)); err != nil {
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
