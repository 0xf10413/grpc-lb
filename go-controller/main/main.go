package main

import (
	"log"
	"net/http"
	"text/template"
	"time"
)

type ServerStatus struct {
	NbClients    int32
	MaxNbClients int32
}

type ServerStatuses = map[string]ServerStatus

const minStableIter = 10

func makeViewStatusHandler(rebalancer *Rebalancer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: move above when status.html stops changing
		t, err := template.ParseFiles("status.html")
		if err != nil {
			log.Panicf("Could not load template: %v", err)
		}
		data := struct {
			HostStatuses        ServerStatuses
			RemainingIterations int
		}{
			rebalancer.LastServerStatuses,
			(rebalancer.MinNbStableIter - rebalancer.NbStableIter),
		}

		t.Execute(w, data)
	}
}

func main() {
	log.Println("Starting…")
	serverAddrs := []string{"192.168.1.12:50052", "192.168.39.101:31045"}
	rebalancer := newRebalancer(serverAddrs, minStableIter)
	done := make(chan bool)
	ticker := time.NewTicker(2 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				rebalancer.run()
			case <-done:
				return
			}
		}
	}()

	http.HandleFunc("/", makeViewStatusHandler(rebalancer))
	log.Fatal(http.ListenAndServe(":8082", nil))
}
