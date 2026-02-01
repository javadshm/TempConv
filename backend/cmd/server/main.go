package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/javadshm/TempConv/backend/gen/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

// server implements the ConvertService
type server struct {
	pb.UnimplementedConvertServiceServer
}

// Convert performs temperature conversion
func (s *server) Convert(ctx context.Context, req *pb.ConvertRequest) (*pb.ConvertResponse, error) {
	if req.FromUnit == req.ToUnit {
		return &pb.ConvertResponse{Value: req.Value}, nil
	}

	var result float64
	switch {
	case req.FromUnit == pb.Unit_CELSIUS && req.ToUnit == pb.Unit_FAHRENHEIT:
		result = req.Value*9/5 + 32
	case req.FromUnit == pb.Unit_FAHRENHEIT && req.ToUnit == pb.Unit_CELSIUS:
		result = (req.Value - 32) * 5 / 9
	default:
		return nil, fmt.Errorf("invalid unit combination: from %s to %s", req.FromUnit, req.ToUnit)
	}

	return &pb.ConvertResponse{Value: result}, nil
}

func main() {
	grpcPort := getEnv("GRPC_PORT", "9090")
	httpPort := getEnv("HTTP_PORT", "8080")

	// Start gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterConvertServiceServer(grpcServer, &server{})
	reflection.Register(grpcServer)

	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", grpcPort, err)
	}

	log.Printf("Starting gRPC server on port %s", grpcPort)
	go func() {
		if err := grpcServer.Serve(grpcLis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Start HTTP gateway
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err = pb.RegisterConvertServiceHandlerFromEndpoint(ctx, mux, "localhost:"+grpcPort, opts)
	if err != nil {
		log.Fatalf("Failed to register gateway: %v", err)
	}

	// Add health endpoint
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/health", healthHandler)
	httpMux.Handle("/", cors(mux))

	httpServer := &http.Server{
		Addr:    ":" + httpPort,
		Handler: httpMux,
	}

	log.Printf("Starting HTTP gateway on port %s", httpPort)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to serve HTTP: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down servers...")
	grpcServer.GracefulStop()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
	log.Println("Servers stopped")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// cors wraps a handler to add CORS headers
func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
