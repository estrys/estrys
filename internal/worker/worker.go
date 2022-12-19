package worker

import (
	"context"

	"github.com/hibiken/asynq"

	"github.com/estrys/estrys/internal/config"
	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/logger"
	"github.com/estrys/estrys/internal/worker/queues"
	"github.com/estrys/estrys/internal/worker/tasks"
	"github.com/estrys/estrys/internal/worker/tasks/handlers"
)

func StartBroker(ctx context.Context) error {
	conf := dic.GetService[config.Config]()
	log := dic.GetService[logger.Logger]()

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: conf.RedisAddress},
		asynq.Config{
			BaseContext: func() context.Context { return ctx },
			LogLevel:    asynq.InfoLevel,
			Logger:      log,
			Queues: map[string]int{
				queues.QueueFollows: 1,
				queues.QueueTweets:  1,
			},
			// Specify how many concurrent workers to use
			//Concurrency: 1,
		},
	)

	mux := asynq.NewServeMux()

	mux.HandleFunc(tasks.TypeAcceptFollow, ErrorHandler(TracingHandler(tasks.HandleAcceptFollow)))
	mux.HandleFunc(tasks.TypeRejectFollow, ErrorHandler(TracingHandler(tasks.HandleRejectFollow)))
	mux.HandleFunc(tasks.TypeSendTweet, ErrorHandler(TracingHandler(handlers.HandleSendTweet)))

	log.Info("Starting worker")

	return srv.Run(mux) //nolint:wrapcheck
}
