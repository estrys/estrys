package client

import "github.com/hibiken/asynq"

//go:generate mockery --name BackgroundWorkerClient
type BackgroundWorkerClient interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

type asynQClient struct {
	client *asynq.Client
}

func NewBackgroundWorkerClient(client *asynq.Client) *asynQClient {
	return &asynQClient{client: client}
}

func (a *asynQClient) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	info, err := a.client.Enqueue(task, opts...)
	return info, err //nolint:wrapcheck
}
