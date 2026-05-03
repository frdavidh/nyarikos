package notifications

import "github.com/hibiken/asynq"

type TaskClient struct {
	client *asynq.Client
}

func NewTaskClient(redisAddr, redisPassword string, redisDB int) *TaskClient {
	return &TaskClient{
		client: asynq.NewClient(asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: redisPassword,
			DB:       redisDB,
		}),
	}
}

func (c *TaskClient) Close() error {
	return c.client.Close()
}

func (c *TaskClient) EnqueueLoginNotification(email string) error {
	payload := &LoginNotificationPayload{Email: email}
	data, err := payload.Marshal()
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeEmailLoginNotification, data)
	_, err = c.client.Enqueue(task)
	return err
}
