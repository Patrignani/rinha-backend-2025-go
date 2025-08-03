package workers

import (
	"context"
	"fmt"
	"log"
	"sync"
)

type QueueWorker struct {
	channel  chan []byte
	fallback [][]byte
	mu       sync.Mutex
}

func NewQueueWorker(buffer int) *QueueWorker {
	return &QueueWorker{
		channel:  make(chan []byte, buffer),
		fallback: [][]byte{},
	}
}

func (q *QueueWorker) Send(msg []byte) {
	select {
	case q.channel <- msg:
	default:
		q.mu.Lock()
		q.fallback = append(q.fallback, msg)
		q.mu.Unlock()
		fmt.Println("Fila cheia, mensagem salva no fallback")
	}
}

func (q *QueueWorker) RetryFallback() {
	q.mu.Lock()
	defer q.mu.Unlock()

	newFallback := q.fallback[:0]

	for _, msg := range q.fallback {
		select {
		case q.channel <- msg:
			fmt.Println("Mensagem reprocessada do fallback")
		default:
			newFallback = append(newFallback, msg)
		}
	}

	q.fallback = newFallback
}

func (q *QueueWorker) Consume(ctx context.Context, workers int, process func(context.Context, []byte) error) {
	var wg sync.WaitGroup
	log.Printf("Consume queue start")
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-q.channel:
					if !ok {
						return
					}
					log.Printf("Consume queue run")
					if err := process(ctx, msg); err != nil {
						fmt.Printf("Erro ao processar mensagem %v\n", err)
					}
				}
			}
		}()
	}

	<-ctx.Done()
	wg.Wait()
}

// func (q *QueueWorker) Consume(ctx context.Context, workers int, process func(context.Context, []byte) error) {
// 	//var wg sync.WaitGroup
// 	sem := make(chan struct{}, workers)

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			//wg.Wait()
// 			close(q.channel)
// 			fmt.Println("Consumo encerrado")
// 			return
// 		case msg, ok := <-q.channel:
// 			if !ok {
// 				//wg.Wait()
// 				fmt.Println("Canal fechado e todas mensagens processadas")
// 				return
// 			}

// 			sem <- struct{}{}
// 			//wg.Add(1)

// 			go func(m []byte) {
// 				//defer wg.Done()
// 				defer func() { <-sem }()
// 				if err := process(ctx, m); err != nil {
// 					fmt.Printf("Erro ao processar mensagem %v\n", err)
// 				}
// 			}(msg)
// 		}
// 	}
// }

func (q *QueueWorker) CountFallback() int {
	return len(q.fallback)
}
