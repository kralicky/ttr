package main

import (
	"fmt"
	"os"

	"github.com/kralicky/ttr/pkg/game"
	"github.com/kralicky/ttr/pkg/ttr"
)

func main() {
	code := make(chan int, 1)
	cmd := ttr.BuildRootCmd()
	go func() {
		err := cmd.Execute()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			code <- 1
		} else {
			code <- 0
		}
		game.ShutdownGLFW()
	}()

	game.RunGLFW()
	os.Exit(<-code)
}
