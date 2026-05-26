package worker

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/rabbitmq/amqp091-go"

	"rwb-contest/internal/dto"
	"rwb-contest/internal/service"
)

type Pool struct {
	workersCount int
	service      *service.TrendsService
}

func NewPool(workersCount int, service *service.TrendsService) *Pool {
	return &Pool{
		workersCount: workersCount,
		service:      service,
	}
}

func (p *Pool) Start(ctx context.Context, jobs <-chan amqp091.Delivery) {
	var wg sync.WaitGroup

	for i := 0; i < p.workersCount; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return

				case msg, ok := <-jobs:
					if !ok {
						return
					}

					p.handleMessage(workerID, msg)
				}
			}
		}(i + 1)
	}

	<-ctx.Done()
	wg.Wait()
}

func (p *Pool) handleMessage(workerID int, msg amqp091.Delivery) {
	var event dto.SearchEvent

	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("worker %d: ошибка разбора сообщения: %v", workerID, err)
		_ = msg.Nack(false, false)
		return
	}

	p.service.ProcessEvent(event)

	if err := msg.Ack(false); err != nil {
		log.Printf("worker %d: ошибка подтверждения сообщения: %v", workerID, err)
		return
	}
}
