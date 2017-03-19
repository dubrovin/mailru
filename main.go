package main

import (
	"fmt"
	"time"
	"github.com/dubrovin/mailru/counter"
	"os"
)


func main() {
	cntr := counter.NewCounter(5, "Go")

	start := time.Now()
	cntr.ScanFile(os.Stdin)
	cntr.Count()
	end := time.Now()

	fmt.Println(end.Sub(start).String())
}