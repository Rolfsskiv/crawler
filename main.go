package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

type Config struct {
	TimeoutLoadPage time.Duration
}

var cfg = Config{
	TimeoutLoadPage: 3,
}
type Error struct {
	Code int
	Message string
}

var client *http.Client

func main() {
	t := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	client = &http.Client{
		Transport: t,
	}

	http.HandleFunc("/sites", search)

	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    35 * time.Second,
		WriteTimeout:   35 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())

}

func search(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		param := r.URL.Query().Get("search")
		res := getContentSite(param)
		if res != nil {
			data, err := json.Marshal(res)
			if err != nil {
				data, _ := json.Marshal(&Error{Code: 400, Message: "Unknown error"})
				w.Write(data)
				// w.WriteHeader(400)
				return
			}
			w.Write(data)
		} else {
			data, _ := json.Marshal(&Error{Code: 200, Message: "Data is nil"})
			w.Write(data)
		}
		// w.WriteHeader(200)

		return
	}
	data, _ := json.Marshal(&Error{Code: 404, Message: "Route not found"})
	w.Write(data)
	// w.WriteHeader(404)
}



func getContentSite(word string) []Result {
	res, err := client.Get(fmt.Sprintf(baseYandexURL, word))
	if err != nil {
		log.Printf("Get yandex err: %s", err.Error())
		return nil
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Printf("Get yandex body err: %s", err.Error())
			return nil
		}
		sites := parseYandexResponse(bodyBytes)
		var urls []string
		for _, site := range sites.Items {
			urls = append(urls, site.Url)
		}
		start := time.Now()
		res := NewBenchmark(urls)
		duration := time.Since(start)
		log.Printf("Time at work: %s", duration)

		return res
	}
	return nil
}