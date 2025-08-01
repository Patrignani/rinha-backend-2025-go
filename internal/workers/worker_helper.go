package workers

import (
	"context"
	"log"
	"time"
)

func StartWorker(ctx context.Context, name string, interval time.Duration, work func(ctx context.Context) error) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		log.Printf("[%s] Worker iniciado (intervalo: %v)", name, interval)

		for {
			select {
			case <-ctx.Done():
				log.Printf("[%s] Encerrando worker", name)
				return
			case <-ticker.C:
				err := work(ctx)
				if err != nil {
					log.Printf("[%s] Erro ao executar tarefa: %v", name, err)
				}
			}
		}
	}()
}
