package counter

import (
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"bufio"
	"os"
	"net/http"
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

func (counter *Counter) ScanFile(f *os.File) {
	counter.wg.Add(1)
	go func() {
		defer counter.wg.Done()
		scanner := bufio.NewScanner(f)
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

func (counter *Counter) GetURLsChanLen() int {
	return len(counter.urls)
}