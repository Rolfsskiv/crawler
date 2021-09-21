package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var cl *http.Client

type Result struct {
	Url   string `json:"url"`
	Count int64  `json:"count"`
}

func NewBenchmark(urls []string) []Result {
	t := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   cfg.TimeoutLoadPage * time.Second,
			KeepAlive: cfg.TimeoutLoadPage * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		TLSHandshakeTimeout:   cfg.TimeoutLoadPage * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	cl = &http.Client{
		Transport: t,
		Timeout:   cfg.TimeoutLoadPage * time.Second,
	}
	result := make(chan Result, 100)
	wg := sync.WaitGroup{}
	for _, u := range urls {
		wg.Add(1)
		worker := newWorker(u, result)
		go worker.Run(&wg)
	}
	log.Printf("Start test sites -> %d", len(urls))
	wg.Wait()
	log.Printf("End test sites -> %d", len(urls))
	res := make([]Result, 0, len(urls))
	for {
		if len(res) == len(urls) {
			break
		}

		r := <-result
		res = append(res, r)
	}

	return res
}

type Worker struct {
	ctx           context.Context
	url           string
	NumGoroutines int64
	end           chan error
	result        chan Result
}

func newWorker(u string, result chan Result) *Worker {
	parseUrl, err := url.Parse(u)
	if err != nil {
		panic(err)
	}

	w := Worker{
		url:           parseUrl.String(),
		NumGoroutines: 0,
		end:           make(chan error),
		ctx:           context.TODO(),
		result:        result,
	}

	return &w
}

func (w *Worker) Run(wg *sync.WaitGroup) {
	defer func() {
		w.result <- Result{
			Url:   w.url,
			Count: w.NumGoroutines,
		}
		wg.Done()
	}()

	go func() {
		for {
			select {
			case <-time.After(time.Second * cfg.TimeoutLoadPage):
				return
			case <-w.end:
				return
			default:
				w.fetch(w.ctx)
			}
		}
	}()
	<-time.After(time.Second * 3)
}

func (w *Worker) fetch(ctx context.Context) {
	req, err := http.NewRequestWithContext(ctx, "GET", strings.ToLower(w.url), nil)
	if err != nil {
		w.end <- err
		return
	}
	res, err := cl.Do(req)
	if err != nil {
		w.end <- err
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		w.end <- errors.New("Code not 200")
		return
	}

	w.NumGoroutines++
}
