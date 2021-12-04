package pool

import "time"

type configuration struct {
	PoolName        string        //the pool name used for logging
	NumberOfWorkers int           //the total number of workers for the pool
	InputTimeout    time.Duration //the timeout when a ticker is used for triggering
}

//Configuration is a decorator function used to get configuration into
// and out of the pool, it can't be used directly since the pool
// pointer isn't exported, but can be used with the New() or Start()
// functions
type Configuration func(p *pool)

func WithInputTimeout(timeout time.Duration) Configuration {
	return func(p *pool) {
		if timeout > 0 {
			p.config.InputTimeout = timeout
		}
	}
}

func WithLogger(logger Logger) Configuration {
	return func(p *pool) {
		p.logger = logger
	}
}

func WithPoolName(poolName string) Configuration {
	return func(p *pool) {
		p.config.PoolName = poolName
	}
}

func WithNumberOfWorkers(numberOfWorkers int) Configuration {
	return func(p *pool) {
		p.config.NumberOfWorkers = numberOfWorkers
	}
}

func WithWorkerFx(workerFx WorkerFx) Configuration {
	return func(p *pool) {
		p.workerFx = workerFx
	}
}

func WithQueueIn(input QueueInput) Configuration {
	return func(p *pool) {
		if input != nil {
			p.input = input
		}
	}
}

func WithQueueOut(output QueueOutput) Configuration {
	return func(p *pool) {
		if output != nil {
			p.output = output
		}
	}
}

func WithInputSignal(signal Signal) Configuration {
	return func(p *pool) {
		if signal != nil {
			p.signal = signal
		}
	}
}
