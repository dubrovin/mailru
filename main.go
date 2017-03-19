package main

import (
	"fmt"
	"os"
	"bufio"

	"strings"
	"net/http"
	"io/ioutil"

	"sync"
	"time"
)

type Counter struct {
	wg            *sync.WaitGroup
	mu            *sync.Mutex
	maxGoroutines chan struct{}
	urls          chan string
	quit          chan struct{}
	total         int
	word          string
}

func NewCounter(maxGoroutines int, word string) *Counter {
	return &Counter{
		wg: &sync.WaitGroup{},
		mu: &sync.Mutex{},
		maxGoroutines: make(chan struct{}, maxGoroutines),
		urls: make(chan string, 1000),
		quit: make(chan struct{}),
		word: word,
		total: 0,
	}
}

func (counter *Counter) ScanStdin() {
	counter.wg.Add(1)
	go func() {
		defer counter.wg.Done()
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			counter.wg.Add(1)
			counter.urls <- scanner.Text()
		}
	}()
}

func (counter *Counter) Count() {
	go func() {
		for {
			select {
			case url := <-counter.urls:
				counter.maxGoroutines <- struct{}{}
				go func(url string) {
					defer counter.wg.Done()
					local_total := 0
					resp, err := http.Get(url)
					if err != nil {
						fmt.Errorf("Error in request %v", err)
						return
					}
					defer resp.Body.Close()
					body, err := ioutil.ReadAll(resp.Body)
					for _, w := range strings.Fields(string(body)) {
						if strings.Contains(w, "Go") {
							local_total++
						}
					}
					fmt.Println(url, local_total)
					counter.mu.Lock()
					counter.total += local_total
					counter.mu.Unlock()
					<-counter.maxGoroutines

				}(url)
			case <-counter.quit:
				return

			}
		}
	}()

	counter.wg.Wait()
	counter.quit <- struct{}{}
	fmt.Println("Total:", counter.total)
}

func main() {
	cntr := NewCounter(5, "Go")
	start := time.Now()
	cntr.ScanStdin()
	cntr.Count()
	end := time.Now()

	fmt.Println(end.Sub(start).String())
}