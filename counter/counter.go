package counter

import (
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"bufio"
	"os"
	"net/http"
	"os/signal"
	"syscall"
	neturl "net/url"
)

type Counter struct {
	wg            *sync.WaitGroup
	mu            *sync.Mutex
	maxGoroutines chan struct{}
	urls          chan string
	quit          chan struct{}
	total         int
	word          string //слово, которое будем искать
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

//ScanFile в отдельной горутине читает урлы и закидывает в канал
func (counter *Counter) ScanFile(f *os.File) {
	stats, err := f.Stat()
	if err != nil {
		fmt.Errorf("error while in getting stats, error: %v", err)
		return
	}
	if (stats.Mode() & os.ModeCharDevice) != 0 {
		f = os.Stdin
		fmt.Println("Urls : ")
	}
	counter.wg.Add(1)
	go func() {
		defer counter.wg.Done()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {

			parsedUrl, err := neturl.Parse(scanner.Text())
			if err != nil {
				fmt.Println("Url error parsing", err)
				return
			}
			counter.wg.Add(1)
			counter.urls <- parsedUrl.String()

		}
	}()
}

//waitForSignal останавливает чтение по сигналу
func (counter *Counter) waitForSignal() {
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			counter.wg.Done()
			counter.quit <- struct{}{}
		}
	}()
}

func (counter *Counter) ReadUrl(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return string(body), nil
}

//Count подсчитываем количество слов counter.word по каждому url
func (counter *Counter) Count() {
	counter.waitForSignal()
	go func() {
		for {
			select {
			case url := <-counter.urls:
				counter.maxGoroutines <- struct{}{}
				go func(url string) {
					defer counter.wg.Done()
					local_total := 0
					body, err := counter.ReadUrl(url)
					if err != nil {
						fmt.Errorf("Error in request %v", err)
						return
					}

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
				break
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