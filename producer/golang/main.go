package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/adhocore/gronx/pkg/tasker"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	SetupTelemetry(ctx)

	taskr := tasker.New(tasker.Option{
		Verbose: true,
		Tz:      "Europe/Berlin",
	})

	taskr.Task("* * * * *", func(ctx context.Context) (int, error) {

		// Roll some random stuff (Subtasks, number of children per subtask, ect.)
		duration := time.Duration(rand.Intn(60)) * time.Second
		childAmount := rand.Intn(10)
		childDepth := rand.Intn(10)
		parallel := rand.Intn(2) == 0
		handleExampleTask(ctx, duration, childAmount, childDepth, parallel)

		return 0, nil
	})

	taskr.Run()
}

func handleExampleTask(ctx context.Context, duration time.Duration, childAmount int, childDepth int, parallel bool) {
	if childDepth <= 0 {
		return
	}

	// Generate a random name for work instead of just using "test"
	taskName := fmt.Sprintf("Task-%d", rand.Intn(1000))
	ctx, span := tracer.Start(ctx, taskName)
	defer span.End()

	start := time.Now()
	end := start.Add(duration)

	// Use 10% of the total time we have to fake initialization work
	logger.Info("Setup started", zap.String("taskname", taskName))
	initDuration := time.Duration(duration.Seconds() * 0.1 * float64(time.Second))
	time.Sleep(initDuration)
	logger.Info("Setup completed", zap.String("taskname", taskName))

	// Run subtasks in goroutines (if parallel, otherwise just in functions) to fake threading/subservices
	logger.Info("Work started", zap.String("taskname", taskName))
	subDuration := time.Duration(((duration.Seconds() * 0.8) / float64(childAmount)) * float64(time.Second))
	if parallel {
		subDuration = time.Duration(duration.Seconds() * 0.8 * float64(time.Second))
		for i := 0; i < childAmount; i++ {
			parallel := rand.Intn(2) == 0
			go handleExampleTask(ctx, subDuration, rand.Intn(childAmount), childDepth-1, parallel)
		}
	} else {
		for i := 0; i < childAmount; i++ {
			parallel := rand.Intn(2) == 0
			handleExampleTask(ctx, subDuration, rand.Intn(childAmount), childDepth-1, parallel)
		}
	}

	// Use the remaining duration we have available for "work"
	workDuration := end.Sub(time.Now())
	time.Sleep(workDuration)
	logger.Info("Work completed", zap.String("taskname", taskName))
}
