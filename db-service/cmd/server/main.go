package main

import (
	"log"
	"net"

	dbProto "db-service/proto"

	"db-service/internal/database"
	"db-service/internal/service"

	"google.golang.org/grpc"
)

//protoc -I ./proto -I . --go_out=. --go-grpc_out=. --grpc-gateway_out=. database.proto

func main() {
	if err := database.InitDB(); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	grpcServer := grpc.NewServer()
	dbService := service.NewDatabaseService(database.GetDB())
	dbProto.RegisterDatabaseServiceServer(grpcServer, dbService)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("Database service is running on port :50052")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
