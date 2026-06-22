package main

import (
	"context"
	"log"
	"net"
	"os"

	flightfinder "github.com/pucora/pucora-grpc-mock/flight_finder"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type flightsServer struct {
	flightfinder.UnimplementedFlightsServer
}

func (flightsServer) FindFlight(_ context.Context, _ *flightfinder.FindFlightRequest) (*flightfinder.FindFlightResponse, error) {
	return &flightfinder.FindFlightResponse{
		Flights: []*flightfinder.Flight{{
			Id:          "FL-001",
			Destination: "NYC",
		}},
	}, nil
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		conn, err := net.Dial("tcp", "localhost:4242")
		if err != nil {
			log.Printf("healthcheck failed: %v", err)
			os.Exit(1)
		}
		conn.Close()
		os.Exit(0)
	}
	addr := envOr("LISTEN_ADDR", ":4242")
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	s := grpc.NewServer()
	flightfinder.RegisterFlightsServer(s, flightsServer{})
	reflection.Register(s)
	log.Printf("mock gRPC backend listening on %s", addr)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
