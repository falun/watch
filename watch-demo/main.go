package main

import (
	"fmt"
	"os"
	"time"

	"github.com/falun/watch"
)

func mainPoll(path string) {
	t1 := time.NewTicker(1 * time.Second).C
	w := watch.File(path)

	go func() {
		for {
			<-t1
			updated, err := w.Updated()
			fmt.Printf("Updated? %v, %v\n", updated, err)
		}
	}()

	time.Sleep(10 * time.Second)
	fmt.Printf("fin\n")
}

func mainEmit(path string) {
	w := watch.File(path)

	ch, cancelFn := w.OnInterval(1 * time.Second)

	go func() {
		for range ch {
			fmt.Printf("Updated!\n")
		}
	}()

	time.Sleep(10 * time.Second)

	fmt.Printf("Cancelling watch interval\n")
	cancelFn()

	time.Sleep(1 * time.Second)
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: watch-demo <poll|emit> <path>\n")
		return
	}

	style := os.Args[1]
	path := os.Args[2]

	if style == "poll" {
		mainPoll(path)
	} else {
		mainEmit(path)
	}
}
