package main

import (
	"fmt"
	"os"
	"bufio"
	//"net/http"
	//"io/ioutil"
	"strings"
	"net/http"
	"io/ioutil"
	//"sync"
	"sync"
)


func main() {
	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)
	k := make(chan struct{}, 5)
	urls := make(chan string, 1000)
	quit := make(chan struct{})
	total := 0


	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			wg.Add(1)
			urls <- scanner.Text()
		}
	}()

	go func() {
		for {
		select {
		case url := <- urls :
			k <- struct {}{}
			go func(url string) {
				defer wg.Done()
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
				mu.Lock()
				total += local_total
				mu.Unlock()

			}(url)
		case <- quit:
			return

		}
	}
	}()

	wg.Wait()
	quit <- struct {}{}
	fmt.Println("Total:", total)
}