package pool

import (
	goqueue "github.com/antonio-alexander/go-queue"
)

const (
	FmtPoolPrefixNamedf              string = "pool[%s]: "
	FmtPoolPrefixUnamedf             string = "pool: "
	ErrPoolStarted                   string = "pool is started"
	ErrPoolNotStarted                string = "pool not started"
	ErrWorkerFxInvalid               string = "workerFx is invalid"
	ErrInvalidNumberOfWorkers        string = "invalid number of workers"
	ErrInputNoInputAndSignal         string = "no input provided and no valid signal"
	InfoPoolStarting                 string = "starting with %d workers..."
	InfoPoolStarted                  string = "started."
	InfoPoolStopping                 string = "stopping..."
	InfoPoolStopped                  string = "stopped."
	InfoWorkerStartedf               string = "worker[%02d] started."
	InfoWorkerFxStartedf             string = "worker[%02d] started fx..."
	InfoWorkerFxFinishedf            string = "worker[%02d] finished fx."
	InfoWorkerOutputOverflowf        string = "worker[%02d] overflow during output enqueue."
	InfoWorkerOutputOverflowInitialf string = "worker[%02d] overflow during output enqueue, will attempt to enqueue until successful or stopped."
	InfoWorkerInputUnderflowf        string = "worker[%02d] underflow during input dequeue."
	InfoWorkerStoppedf               string = "worker[%02d] stopped."
	InfoWorkerStoppedOutputEnqueuef  string = "worker[%02d] stopped during output enqueue."
	InfoWorkerStartingf              string = "worker[%02d] starting..."
	InfoWorkerStoppingf              string = "worker[%02d] stopping..."
	InfoPanicf                       string = "worker[%02d] experienced panic: %#v"
)

type Signal <-chan struct{}

type WorkerFx func(workerID int, input interface{}) (output interface{})

type QueueInput interface {
	goqueue.Dequeuer
	goqueue.Event
}

type QueueOutput interface {
	goqueue.Enqueuer
	goqueue.EnqueueInFronter
	goqueue.Event
}

type Logger interface {
	Printf(format string, v ...interface{})
}

type Pool interface {
	Start(configs ...Configuration) (err error)
	Stop() (err error)
}
