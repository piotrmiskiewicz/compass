package process

import (
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	"github.com/sirupsen/logrus"
)

const (
	workersAmount = 5
)

type Executor interface {
	Execute(operationID string) (time.Duration, error)
}

type Queue struct {
	queue    workqueue.RateLimitingInterface
	executor Executor
}

func NewQueue(executor Executor) *Queue {
	return &Queue{
		queue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "operations"),
		executor: executor,
	}
}

func (q *Queue) Add(processId string) {
	q.queue.Add(processId)
}

func (q *Queue) Run(stop <-chan struct{}) {
	var waitGroup sync.WaitGroup

	for i := 0; i < workersAmount; i++ {
		createWorker(q.queue, q.executor.Execute, stop, &waitGroup)
	}
}

func createWorker(queue workqueue.RateLimitingInterface, process func(id string) (time.Duration, error), stopCh <-chan struct{}, waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)
	go func() {
		wait.Until(worker(queue, process), time.Second, stopCh)
		waitGroup.Done()
		logrus.Error("Worker done")
	}()
}

func worker(queue workqueue.RateLimitingInterface, process func(key string) (time.Duration, error)) func() {
	return func() {
		exit := false
		for !exit {
			exit = func() bool {
				key, quit := queue.Get()
				if quit {
					logrus.Warnf("Queue end!!!")
					return true
				}
				defer func() {
					r := recover()
					if r != nil {
						logrus.Errorf("Panic handled: %s", r)
					}
					queue.Done(key)
					logrus.Infof("Queue.len %d", queue.NumRequeues(key))


				}()


				when, err := process(key.(string))
				logrus.Infof("When: %s", when)
				if err == nil && when != 0 {
					logrus.Infof("Adding after %s %s", key, when)
					queue.AddAfter(key, when)
					return false
				}

				queue.Forget(key)
				return false
			}()
		}
	}
}
