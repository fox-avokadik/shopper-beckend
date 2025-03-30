package main

import (
	"auth-service/internal/middleware"
	authProto "auth-service/proto"
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"auth-service/internal/handler"
	"auth-service/internal/repository"
	"auth-service/internal/service"
	dbProto "db-service/proto"
)

//protoc -I ./proto -I . --go_out=. --go-grpc_out=. --grpc-gateway_out=. auth.proto

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "localhost:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("Failed to connect to Database service: %v", err)
	}
	defer conn.Close()

	dbClient := dbProto.NewDatabaseServiceClient(conn)

	userRepo := repository.NewUserRepository(dbClient)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	authInterceptor := middleware.NewAuthInterceptor(userRepo)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
	)
	authProto.RegisterAuthServiceServer(grpcServer, userHandler)

	listener, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("Failed to listen on port 50053: %v", err)
	}

	gwMux := runtime.NewServeMux(
		runtime.WithErrorHandler(handler.CustomErrorHandler),
	)
	err = authProto.RegisterAuthServiceHandlerFromEndpoint(
		context.Background(),
		gwMux,
		"localhost:50053",
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	)
	if err != nil {
		log.Fatalf("Failed to register HTTP gateway: %v", err)
	}

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: gwMux,
	}

	go func() {
		log.Println("Starting gRPC server on :50053...")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v", err)
		}
	}()

	log.Println("Starting HTTP server on :8080...")
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
