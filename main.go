package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type ServerList struct {
	URL     string
	Healthy bool
}
type ServerPool struct {
	servers []ServerList
	mux     sync.Mutex
}

func httpHandler(nextHop func() string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		urlPtr, _ := url.Parse("http://" + nextHop())
		revPxy := httputil.NewSingleHostReverseProxy(urlPtr)
		revPxy.ServeHTTP(w, r)
	}
}

func roundRobin(list *ServerPool) func() string {
	var counter int
	return func() string {
		list.mux.Lock()
		defer list.mux.Unlock()

		serverLen := len(list.servers)
		if serverLen == 0 {
			return "0.0.0.0:80"
		}
		for i := 0; i < serverLen; i++ {
			//fmt.Printf("is %v healthy? - %v\n", list.servers[i], list.servers[i].Healthy)

			idx := counter % serverLen
			counter++
			if !(list.servers[idx].Healthy) {
				continue
			}
			return list.servers[idx].URL
		}
		return "0.0.0.0:80"
	}
}
func tmpHandler(string2 string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(writer, "<h1>Hey its your pal from %s<h1>", string2)
	}
}

func healthHandler(string2 string, num int) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(time.Second * time.Duration(num))
		fmt.Fprintf(writer, "<h1>Hey im healthy %s<h1>", string2)
	}
}

func healthWorker(list *ServerPool) {
	for {
		time.Sleep(time.Second * 5)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		done := make(chan struct{}, 1)
		num := rand.Intn(len(list.servers))
		go func(ctx context.Context, num int, list *ServerPool, done chan<- struct{}) {
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+list.servers[num].URL+"/health", nil)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				list.mux.Lock()
				defer list.mux.Unlock()
				list.servers[num].Healthy = true
				done <- struct{}{}
			}
			return

		}(ctx, num, list, done)

		select {
		case <-done:
			cancel()
			continue
		case <-ctx.Done():
			list.mux.Lock()
			list.servers[num].Healthy = false
			list.mux.Unlock()
			cancel()
			continue
		}
	}
}
func main() {
	serverRange := ServerPool{}
	serverRange.servers = make([]ServerList, 0)
	for i := range 6 {
		go func(num int) {
			lbMux := http.NewServeMux()
			host := fmt.Sprintf("127.0.0.1:808%v", num)
			serverRange.mux.Lock()
			serverRange.servers = append(serverRange.servers, ServerList{host, false})
			serverRange.mux.Unlock()
			lbMux.HandleFunc("/health", healthHandler(host, num))
			lbMux.HandleFunc("/", tmpHandler(host))
			err := http.ListenAndServe(host, lbMux)
			if err != nil {
				log.Fatalln(err)
			}
		}(i + 1)

	}
	go healthWorker(&serverRange)
	lbMux := http.NewServeMux()
	lbMux.HandleFunc("/", httpHandler(roundRobin(&serverRange)))
	err := http.ListenAndServe("127.0.0.1:8080", lbMux)
	if err != nil {
		log.Fatalln(err)
	}

}
