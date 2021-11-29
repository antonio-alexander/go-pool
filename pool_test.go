package pool_test

import (
	"errors"
	"math/rand"
	"sync"
	"testing"
	"time"

	pool "github.com/antonio-alexander/go-pool"
	finite "github.com/antonio-alexander/go-queue/finite"

	internal_example "github.com/antonio-alexander/go-pool/internal/example"
	internal_logger "github.com/antonio-alexander/go-pool/internal/logger"

	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	var p pool.Pool

	input := finite.New(1)
	output := finite.New(1)
	// start valid
	p = pool.New(
		pool.WithWorkerFx(internal_example.WorkerFxIncrement),
		pool.WithQueueIn(input),
		pool.WithQueueOut(output),
		pool.WithNumberOfWorkers(1),
	)
	// start again
	err := p.Start()
	if assert.NotNil(t, err) {
		assert.Equal(t, errors.New(pool.ErrPoolStarted), err)
	}
	err = p.Stop()
	assert.Nil(t, err)
	// start with panic
	assert.Panics(t, func() {
		p = pool.New(pool.WithNumberOfWorkers(0))
	})
	p = pool.New()
	err = p.Stop()
	assert.NotNil(t, err)
	input.Close()
	output.Close()
}

func TestValidation(t *testing.T) {
	input := finite.New(1)
	output := finite.New(1)
	signal := make(chan struct{})
	closedSignal := make(chan struct{})
	close(closedSignal)
	cases := map[string]struct {
		iConfigs []pool.Configuration
		oErr     error
	}{
		"valid": {
			iConfigs: []pool.Configuration{
				pool.WithWorkerFx(internal_example.WorkerFxIncrement),
				pool.WithQueueIn(input),
				pool.WithQueueOut(output),
				pool.WithNumberOfWorkers(1),
			},
		},
		"no_input_and_no_signal": {
			iConfigs: []pool.Configuration{
				pool.WithWorkerFx(internal_example.WorkerFxIncrement),
				pool.WithQueueOut(output),
				pool.WithNumberOfWorkers(1),
			},
			oErr: errors.New(pool.ErrInputNoInputAndSignal),
		},
		"no_input_and_closed_signal": {
			iConfigs: []pool.Configuration{
				pool.WithWorkerFx(internal_example.WorkerFxIncrement),
				pool.WithQueueOut(output),
				pool.WithNumberOfWorkers(1),
				pool.WithInputSignal(closedSignal),
			},
			oErr: errors.New(pool.ErrInputNoInputAndSignal),
		},
		"zero_number_of_workers": {
			iConfigs: []pool.Configuration{
				pool.WithWorkerFx(internal_example.WorkerFxIncrement),
				pool.WithQueueIn(input),
				pool.WithQueueOut(output),
				pool.WithNumberOfWorkers(0),
			},
			oErr: errors.New(pool.ErrInvalidNumberOfWorkers),
		},
		"invalid_workerFx": {
			iConfigs: []pool.Configuration{
				pool.WithWorkerFx(nil),
				pool.WithQueueIn(input),
				pool.WithQueueOut(output),
				pool.WithNumberOfWorkers(1),
			},
			oErr: errors.New(pool.ErrWorkerFxInvalid),
		},
	}
	for name, c := range cases {
		p := pool.New()
		err := p.Start(c.iConfigs...)
		if c.oErr == nil {
			assert.Nil(t, err, "case: %s", name)
			err = p.Stop()
			assert.Nil(t, err, "case: %s", name)
			continue
		}
		assert.Equal(t, c.oErr, err, "case: %s", name)
	}
	close(signal)
	input.Close()
	output.Close()
}

func TestWorkerFx(t *testing.T) {
	var wg sync.WaitGroup
	var p pool.Pool
	var err error

	//TODO: input with regular output
	input := finite.New(1)
	output := finite.New(1)
	p = pool.New(
		pool.WithNumberOfWorkers(1),
		pool.WithQueueIn(input),
		pool.WithQueueOut(output),
		pool.WithWorkerFx(internal_example.WorkerFxIncrement),
		// pool.WithLogger(&internal_logger.Logger{}),
	)
	assert.Nil(t, err)
	signalOut := output.GetSignalIn()
	overflow := input.Enqueue(1)
	assert.False(t, overflow)
	<-signalOut
	item, underflow := output.Dequeue()
	assert.False(t, underflow)
	valueInt, ok := item.(int)
	assert.True(t, ok, "invalid type: %T", item)
	assert.Equal(t, 2, valueInt)
	err = p.Stop()
	assert.Nil(t, err)
	input.Close()
	output.Close()

	//TODO: input with overflow
	input = finite.New(1)
	output = finite.New(1)
	p = pool.New(
		pool.WithNumberOfWorkers(1),
		pool.WithQueueIn(input),
		pool.WithQueueOut(output),
		pool.WithWorkerFx(internal_example.WorkerFxIncrement),
		pool.WithLogger(&internal_logger.Logger{}),
	)
	assert.Nil(t, err)
	signalIn := input.GetSignalOut()
	signalOut = output.GetSignalIn()
	overflow = input.Enqueue(1)
	assert.False(t, overflow)
	<-signalIn
	underflow = input.Enqueue(2)
	assert.False(t, underflow)
	time.Sleep(time.Second) //this tests that the workerFx will continue to enqueue to output while full
	<-signalOut
	item, underflow = output.Dequeue()
	assert.False(t, underflow)
	valueInt, _ = item.(int)
	assert.Equal(t, 2, valueInt)
	<-signalOut
	item, underflow = output.Dequeue()
	assert.False(t, underflow)
	valueInt, _ = item.(int)
	assert.Equal(t, 3, valueInt)
	err = p.Stop()
	assert.Nil(t, err)
	input.Close()
	output.Close()

	//TODO: no input with signal output
	signal := make(chan struct{})
	output = finite.New(1)
	p = pool.New(
		pool.WithNumberOfWorkers(1),
		pool.WithQueueOut(output),
		pool.WithWorkerFx(internal_example.WorkerFxHelloWorld),
		pool.WithInputSignal(signal),
		// pool.WithLogger(&internal_logger.Logger{}),
	)
	signal <- struct{}{}
	<-output.GetSignalIn()
	item, underflow = output.Dequeue()
	assert.False(t, underflow)
	valueString, ok := item.(string)
	assert.True(t, ok, "type doesn't match")
	assert.Equal(t, "Hello, World!", valueString)
	err = p.Stop()
	assert.Nil(t, err)
	close(signal)
	output.Close()

	//TODO: test input dequeue
	input = finite.New(1)
	p = pool.New(pool.WithNumberOfWorkers(1),
		pool.WithQueueIn(input),
		pool.WithWorkerFx(internal_example.WorkerFxHelloWorld),
		pool.WithInputTimeout(time.Microsecond),
		pool.WithLogger(&internal_logger.Logger{}),
	)
	time.Sleep(2 * time.Microsecond)
	//KIM: this is only meant to trigger the dequeue
	// code path, there's no "functionality" done
	err = p.Stop()
	assert.Nil(t, err)
	input.Close()

	//TODO: workerFx with panic
	signal = make(chan struct{})
	p = pool.New(
		pool.WithPoolName("panic"),
		pool.WithNumberOfWorkers(1),
		pool.WithWorkerFx(internal_example.WorkerFxPanic),
		pool.WithInputSignal(signal),
		pool.WithLogger(&internal_logger.Logger{}),
	)
	signal <- struct{}{}
	time.Sleep(time.Microsecond)
	//KIM: this is only to ensure that the panic code
	// path actually executes, there's no real "functionality"
	// here
	err = p.Stop()
	assert.Nil(t, err)
	close(signal)

	//TODO: workerFx stop while waiting to output
	input = finite.New(1)
	output = finite.New(1)
	p = pool.New(
		pool.WithNumberOfWorkers(1),
		pool.WithQueueIn(input),
		pool.WithQueueOut(output),
		pool.WithWorkerFx(internal_example.WorkerFxIncrement),
		pool.WithLogger(&internal_logger.Logger{}),
	)
	assert.Nil(t, err)
	signalIn = input.GetSignalOut()
	overflow = input.Enqueue(1)
	assert.False(t, overflow)
	<-signalIn
	underflow = input.Enqueue(2)
	assert.False(t, underflow)
	time.Sleep(time.Second) //this tests that the workerFx will continue to enqueue to output while full
	//KIM: he're we're just confirming that we get a different "log" when we stop while waiting
	// to enqueue to output
	err = p.Stop()
	assert.Nil(t, err)
	input.Close()
	output.Close()

	//TODO: output queue is nil
	chData := make(chan int)
	p = pool.New(
		pool.WithNumberOfWorkers(1),
		pool.WithWorkerFx(func(workerID int, input interface{}) (output interface{}) {
			chData <- rand.Int()
			return nil
		}),
		pool.WithInputTimeout(time.Microsecond),
		// pool.WithLogger(&internal_logger.Logger{}),
	)
	wg.Add(1)
	go func() {
		defer wg.Done()

		for valueInt := range chData {
			assert.Greater(t, valueInt, 0)
		}
	}()
	time.Sleep(2 * time.Microsecond)
	err = p.Stop()
	assert.Nil(t, err)
	close(chData)
}

func TestConcurrency(t *testing.T) {
	var wg sync.WaitGroup

	//the goal of this test is to have four different go routines
	// running at the same time and interacting with the pool and
	// the queues simultaneously
	input := finite.New(100)
	output := finite.New(100)
	p := pool.New(
		pool.WithNumberOfWorkers(4),
		pool.WithQueueIn(input),
		pool.WithQueueOut(output),
		pool.WithWorkerFx(internal_example.WorkerFxIncrement),
		// pool.WithLogger(&internal_logger.Logger{}),
	)
	stopperInput := make(chan struct{})
	stopperOutput := make(chan struct{})
	started := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()

		tInput := time.NewTicker(time.Millisecond)
		defer tInput.Stop()
		close(started)
		for {
			select {
			case <-tInput.C:
				if overflow := input.Enqueue(rand.Int()); overflow {
					t.Logf("%s, experienced overflow", t.Name())
				}
			case <-stopperInput:
				return
			}
		}
	}()
	<-started
	started = make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()

		tInput := time.NewTicker(time.Millisecond)
		defer tInput.Stop()
		close(started)
		for {
			select {
			case <-tInput.C:
				if overflow := input.Enqueue(rand.Int()); overflow {
					t.Logf("%s, experienced overflow", t.Name())
				}
			case <-stopperInput:
				return
			}
		}
	}()
	<-started
	started = make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()

		signal := output.GetSignalIn()
		close(started)
		for {
			select {
			case <-signal:
				if _, underflow := output.Dequeue(); underflow {
					t.Logf("%s, experienced underflow", t.Name())
				}
			case <-stopperOutput:
				return
			}
		}
	}()
	<-started
	started = make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()

		signal := output.GetSignalIn()
		close(started)
		for {
			select {
			case <-signal:
				if _, underflow := output.Dequeue(); underflow {
					t.Logf("%s, experienced underflow", t.Name())
				}
			case <-stopperOutput:
				return
			}
		}
	}()
	<-started
	time.Sleep(5 * time.Second)
	close(stopperInput)
	for l := input.Length(); l > 0; {
		l = input.Length() //drains the input
	}
	err := p.Stop()
	assert.Nil(t, err)
	close(stopperOutput)
	wg.Wait()
	input.Close()
	output.Close()
}
