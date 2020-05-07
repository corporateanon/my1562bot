package priorityChecker

import (
	"github.com/hibiken/asynq"
	"github.com/my1562/queue"
	"github.com/my1562/telegrambot/pkg/config"
)

type PriorityChecker struct {
	client *asynq.Client
}

func NewPriorityChecker(config *config.Config) *PriorityChecker {
	redis := asynq.RedisClientOpt{Addr: config.Redis}
	client := asynq.NewClient(redis)

	return &PriorityChecker{
		client: client,
	}
}

func (p *PriorityChecker) EnqueuePriorityCheck(addrID int64) error {
	if err := p.client.Enqueue(queue.NewPriorityCheckTask(addrID)); err != nil {
		return err
	}
	return nil
}
