/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"go.etcd.io/etcd/client/v3/concurrency"
	"go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v Gatech", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName() + " gatech"}, nil
}

func main() {
	var name = flag.String("name", "", "give a name")
	flag.Parse()

	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{"localhost:2379"}})

	if err != nil {
		log.Fatal(err)
	}

	defer cli.Close()

	s, err := concurrency.NewSession(cli)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	e := concurrency.NewElection(s, "/leader-election/")
	ctx := context.Background()

	if err := e.Campaign(ctx, "e"); err != nil {
		log.Fatal(err)
	}

	fmt.Println("leader election for ", *name)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s1 := grpc.NewServer()
	pb.RegisterGreeterServer(s1, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s1.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	if err := e.Resign(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("resign ", *name)
}

