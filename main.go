package main

import (
	"os"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) > 0 {
		increment, _ := strconv.Atoi(os.Args[1])
		time.Sleep(time.Duration(increment) * time.Second)
	}
}
