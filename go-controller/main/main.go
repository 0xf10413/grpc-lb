package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

func retrieveData() (map[string]int32, error) {
	serverAddrs := []string{"localhost:50051", "192.168.39.101:31044"}
	resultMap := make(map[string]int32)

	for _, serverAddr := range serverAddrs {
		log.Printf("About to connect to %s", serverAddr)

		var opts []grpc.DialOption
		opts = append(opts, grpc.WithInsecure())
		conn, err := grpc.Dial(serverAddr, opts...)
		if err != nil {
			log.Fatal("Cannot connect, got error {}", err)
			return nil, err
		}
		defer conn.Close()

		log.Println("Connected!")

		client := NewLoadBalancingManagerClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		reply, err := client.GetClientStatus(ctx, &ClientRequest{})

		if err != nil {
			log.Fatalf("Could not get reply: %v", err)
			return nil, err
		}
		log.Printf("Got reply %v", reply)
		resultMap[serverAddr] = reply.GetNbClients()
	}
	return resultMap, nil
}

func main() {
	log.Println("Starting…")
	done := make(chan bool)
	ticker := time.NewTicker(2 * time.Second)

	/*go*/
	func() {
		for {
			select {
			case <-ticker.C:
				log.Printf("Fetching server data…")
				resultMap, err := retrieveData()
				if err != nil {
					log.Fatalf("Could not get all data: %v", err)
				} else {
					log.Printf("All data = %v", resultMap)
				}
			case <-done:
				return
			}
		}
	}()

	time.Sleep(2 * time.Second)
}
