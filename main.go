package main

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/Patrignani/patrignani-rinha-backend-go/internal/repositories"
	"github.com/Patrignani/patrignani-rinha-backend-go/internal/services"
	"github.com/Patrignani/patrignani-rinha-backend-go/internal/workers"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/config"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/storage"
	"github.com/Patrignani/patrignani-rinha-backend-go/servers"
	"github.com/panjf2000/gnet/v2"
)

func main() {

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic capturado: %v\nStack trace:\n%s", r, debug.Stack())
			panic(r)
		}
	}()

	ctx := context.Background()

	pg, err := storage.NewPostgresClient(ctx, getPostgresDSN())
	if err != nil {
		log.Printf("Caiu panic Postgres")
		panic(fmt.Errorf("erro ao iniciar o banco: %w", err))
	}
	defer pg.Close()

	paymentRepo := repositories.NewPaymentRepository(pg)
	queue := workers.NewQueueWorker(config.Env.Queue.Buffer)
	paymentService := services.NewPaymentService(paymentRepo, queue)

	go queue.Consume(ctx, config.Env.Queue.Workers, paymentService.RunQueue)
	workers.StartWorker(ctx, "Retry", 300*time.Millisecond, func(ctx context.Context) error {
		if queue.CountFallback() > 0 {
			queue.RetryFallback()
		}
		return nil
	})

	server := servers.NewGNetServer(paymentService, true, queue)

	log.Printf(`
	╔════════════════════════════════════════════════════╗
	║ 🚀 Servidor iniciado com sucesso!                   ║
	║ 📡 Escutando em: http://%s                          ║
	╚════════════════════════════════════════════════════╝
	`, fmt.Sprintf(":%s", config.Env.StartPort))

	if err := gnet.Run(server, fmt.Sprintf("tcp://:%s", config.Env.StartPort),
		gnet.WithMulticore(true),
		//	gnet.WithLogger(nil),
		gnet.WithTCPNoDelay(gnet.TCPNoDelay)); err != nil {
		log.Printf("Caiu panic Gnet")
		panic(fmt.Errorf("gnet.Run falhou: %w", err))
	}

	log.Printf("RUN GNET CONCLUIDO")

}

func getPostgresDSN() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", config.Env.Postgres.User, config.Env.Postgres.Pass, config.Env.Postgres.Host, config.Env.Postgres.PORT, config.Env.Postgres.Name)
}
