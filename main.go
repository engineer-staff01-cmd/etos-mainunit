package main

import (
	"embed"
	"etos-mainunit/app"
	"os"
	"runtime"
)

/*
プログラムを強制終了
予期されるエラーに対して正常終了させたい場合などに使われる
*/

//go:embed public/*
var fs_public embed.FS

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)
	os.Exit(app.Run(fs_public))
}
