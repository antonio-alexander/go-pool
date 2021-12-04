package pool

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type pool struct {
	sync.RWMutex
	sync.WaitGroup
	config   configuration
	workerFx WorkerFx
	input    QueueInput
	output   QueueOutput
	signal   Signal
	logger   Logger
	stopper  chan struct{}
	started  bool
}

func New(configs ...Configuration) interface {
	Pool
} {
	p := &pool{stopper: make(chan struct{})}
	if len(configs) > 0 {
		if err := p.Start(configs...); err != nil {
			panic(err)
		}
	}
	return p
}

func (p *pool) launchWorker(workerID int) {
	started := make(chan struct{})
	p.Add(1)
	go func() {
		defer p.Done()
		defer p.printf(InfoWorkerStoppedf, workerID)

		var signal Signal

		workerFx := func() {
			defer func() {
				if r := recover(); r != nil {
					p.printf(InfoPanicf, workerID, r)
				}
			}()

			var input interface{} = nil

			p.printf(InfoWorkerFxStartedf, workerID)
			if p.input != nil {
				item, underflow := p.input.Dequeue()
				if underflow {
					p.printf(InfoWorkerInputUnderflowf, workerID)
					return
				}
				input = item
			}
			output := p.workerFx(workerID, input)
			if p.output == nil {
				return
			}
			signal := p.output.GetSignalOut()
			if overflow := p.output.Enqueue(output); overflow {
				p.printf(InfoWorkerOutputOverflowInitialf, workerID)
				for overflow {
					select {
					case <-p.stopper:
						p.printf(InfoWorkerStoppedOutputEnqueuef, workerID)
						p.printf(InfoWorkerStoppedf, workerID)
						return
					case <-signal:
						if overflow = p.output.Enqueue(output); overflow {
							//KIM: this is a valid case, but is difficult to re-create
							// unless we use an implementation where the signal for the
							// queue isn't unique or implemented correctly
							p.printf(InfoWorkerOutputOverflowf, workerID)
						}
					}
				}
			}
			p.printf(InfoWorkerFxFinishedf, workerID)
		}
		if p.input != nil {
			signal = p.input.GetSignalIn()
		} else {
			signal = p.signal
		}
		tDequeue := time.NewTicker(time.Second)
		tDequeue.Stop()
		if timeout := p.config.InputTimeout; timeout > 0 {
			tDequeue = time.NewTicker(timeout)
			defer tDequeue.Stop()
		}
		close(started)
		p.printf(InfoWorkerStartedf, workerID)
		for {
			select {
			case <-tDequeue.C:
				workerFx()
			case <-signal:
				workerFx()
			case <-p.stopper:
				return
			}
		}
	}()
	<-started
}

func (p *pool) validate() error {
	if p.workerFx == nil {
		return errors.New(ErrWorkerFxInvalid)
	}
	if p.config.NumberOfWorkers <= 0 {
		return errors.New(ErrInvalidNumberOfWorkers)
	}
	if p.input == nil {
		validSignal := false

		if p.config.InputTimeout > 0 {
			validSignal = true
		}
		if !validSignal && p.signal != nil {
			select {
			default:
				validSignal = true
			case <-p.signal:
			}
		}
		if !validSignal {
			return errors.New(ErrInputNoInputAndSignal)
		}
	}
	return nil
}

func (p *pool) printf(format string, v ...interface{}) {
	if logger := p.logger; logger != nil {
		if p.config.PoolName != "" {
			prefix := fmt.Sprintf(FmtPoolPrefixNamedf, p.config.PoolName)
			logger.Printf(prefix+format+"\n", v...)
		} else {
			logger.Printf(FmtPoolPrefixUnamedf+format+"\n", v...)
		}
	}
}

func (p *pool) Start(configs ...Configuration) (err error) {
	p.Lock()
	defer p.Unlock()

	if p.started {
		return errors.New(ErrPoolStarted)
	}
	for _, config := range configs {
		config(p)
	}
	if err = p.validate(); err != nil {
		return
	}
	p.printf(InfoPoolStarting, p.config.NumberOfWorkers)
	for i := 0; i < p.config.NumberOfWorkers; i++ {
		workerID := i + 1
		p.printf(InfoWorkerStartingf, workerID)
		p.launchWorker(workerID)
	}
	p.started = true
	p.printf(InfoPoolStarted)

	return
}

func (p *pool) Stop() (err error) {
	p.Lock()
	defer p.Unlock()

	if !p.started {
		return errors.New(ErrPoolNotStarted)
	}
	p.printf(InfoPoolStopping)
	close(p.stopper)
	p.Wait()
	p.printf(InfoPoolStopped)

	return
}
