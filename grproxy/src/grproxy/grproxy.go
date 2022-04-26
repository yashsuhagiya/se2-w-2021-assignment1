package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

var listofURL []string

func main() {

	server := make([]string, 1)
	server[0] = "zookeeper:2181"
	conn, _, err := zk.Connect(server, time.Second)
	showError(err)
	defer conn.Close()

	for conn.State() != zk.StateHasSession {
		fmt.Printf("Loading Zookeeper from grproxy..\n")
		second := time.Second
		time.Sleep(15 * second)
	}

	exists, _, err := conn.Exists("/grproxy")
	showError(err)

	if !exists {
		grproxy, err := conn.Create("/grproxy", []byte("grproxy:80"), int32(0), zk.WorldACL(zk.PermAll))
		showError(err)
		fmt.Printf("CREATE: %+v\n", grproxy)
	}

	chnchild := make(chan []string)
	errors := make(chan error)
	go func() {
		for {
			child, _, events, err := conn.ChildrenW("/grproxy")
			if err != nil {
				errors <- err
				return
			}
			chnchild <- child
			evnt := <-events
			if evnt.Err != nil {
				errors <- evnt.Err
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case child := <-chnchild:
				fmt.Printf("%+v \n", child)
				var temp []string
				for _, child := range child {
					gserveUrlList, _, err := conn.Get("/grproxy/" + child)
					temp = append(temp, string(gserveUrlList))
					if err != nil {
						fmt.Printf("CHILD ERROR: %+v\n", err)
					}
				}
				listofURL = temp
				fmt.Printf("%+v \n", listofURL)
			case err := <-errors:
				fmt.Printf("OTHER ERRORS: %+v\n", err)
			}
		}
	}()

	proxy := NewMultipleHostReverseProxy()
	log.Fatal(http.ListenAndServe(":8080", proxy))
}

func NewMultipleHostReverseProxy() *httputil.ReverseProxy {

	director := func(req *http.Request) {

		if req.URL.Path == "/library" {
			fmt.Println("This is handled by gserver....")
			hostName := listofURL[rand.Int()%len(listofURL)]
			req.URL.Host = hostName
			req.URL.Scheme = "http"

		} else {
			fmt.Println("This is handled by nginx....")
			req.URL.Scheme = "http"
			req.URL.Host = "nginx"
		}

	}
	return &httputil.ReverseProxy{Director: director}
}

func showError(err error) {
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
	}
}
