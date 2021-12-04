# github.com/antonio-alexander/go-pool

go-pool provides a library to simplify long-lived worker pools. The expected use case for this library is NOT for short term, locally-lived pools, these are operations that are long-lived and can benefit from go routines. Generally, this library abstracts common functionality/actions that you otherwise would need to duplicate to produce the same functionality.

go-pool benefits heavily when used in producer/consumer patterns with the ability to "tune" by setting the size of the input/output queues, the number of workers as well as how the workers are triggered. go-pool heavily uses the [go-queue](http://github.com/antonio-alexander/go-queue) interfaces to get data in and out of the queue. The following use cases are supported (implementation details are below):

- an input and output queue, with the worker (callback) function executed when data is placed into the input queue (using the SignalIn() function)
- no input queue (with or without an output queue) and the workerFx triggered via a time.Ticker
- no input queue or output queue and the workerFx triggered via time.Ticker
- no input queue (with or without an output queue) and the workerFx triggered via an external signal
- no input queue or output queue and the workerFx triggered via an external signal
- a nil workerFx

## Getting Started

To get started you'll need to create an instance of the pool using the New() function and then provide a valid configuration to start the queue. The New() function has a variatic that will attempt to start the pool (or panic). See the configuration section for valid configuration options.

This example will create a pool with a single worker, a configured input and ouput queue as well as a workerFx that ignores it's input and only provides an output that prints "Hello, World!".

```go
package main

import (
    pool "github.com/antonio-alexander/go-pool"
    finite "github.com/antonio-alexander/go-queue/finite"
)

func workerFx(workerID int, input interface{}) (output interface{}) {
    return "Hello, World!"
}

func main(){
    input := finite.New(1)
    output := finite.New(1)
    p := pool.New()
    err := p.Start(
        pool.WithNumberOfWorkers(1),
        pool.WithQueueIn(input),
        pool.WithQueueOut(output),
        pool.WithWorkerFx(workerFx),
    )
    if err != nil {
        fmt.Println(err)
        return
    }
    signalOut := output.GetSignalIn()
    overflow := input.Enqueue(1)
    assert.False(t, overflow)
    <-signalOut
    item, underflow := output.Dequeue()
    assert.False(t, underflow)
    value, _ := item.(string)
    fmt.Println(value)
    err = p.Stop()
    if err != nil {
        fmt.Println(err)
    }
    input.Close()
    output.Close()
}
```

```sh
Hello, World!
```

## Configuration

There are a handful of configuration options for the pool (with some invalid combinations):

- WithInputTimeout(timeout time.Duration) - can be used to create a time-based signal (when no input queue is provided)
- WithLogger(logger Logger) - can be used to pass through a logger as long as whatever's provided implements Logger
- WithPoolName(poolName string) - can be used to provide a pool name that's used with the logger (otherwise it's useless)
- WithNumberOfWorkers(numberOfWorkers int) - can be used to configure the number of workers created on pool start
- WithWorkerFx(workerFx WorkerFx) - is used to provide the workerFx to be used when the pool is signalled
- WithQueueIn(input QueueInput) - is used to provide an input data structure
- WithQueueOut(output QueueOutput) - is used to provide an output data structure
- WithInputSignal(signal Signal) - is used to provide external signalling when no input is provided (or you want to ignore the signal from the input queue)

These are incompatible configurations (will illicit a panic or error on start:

- no input queue with no external signal or timeout (there's nothing to trigger the workerFx)
- no input queue or external signal with a closed or nil external signal
- a number of workers that's less or equal to zero

## WorkerFx

The workerFx is the function that's executed each time a worker is triggered, the workerFx will have the workerID available as well as the input and output (as empty interfaces). When developing workerFx code, be careful such that they don't block unnecessarily and that they return output values if expected. The workerFx is wrapped in a recover() function that should capture and log any panics (if you have the logger configured). The workerFx has a direct affect on your throughput, so the more efficient it is, the higher your throughput.

```go
type WorkerFx func(workerID int, input interface{}) (output interface{})
```

## Throughput/Tuning

Although most of the throughput/tuning goes hand-in-hand with the producer/consumer design pattern, pool provides an additional item that can be tuned: the number of workers. The total number of "units of work" that can be in-flight using finite resources is: **the size of the input queue + the number of workers + size of the output queue**. Thus with an input and output queue size of 10 and 10 workers, you'd have a total of 30 units of work that can be in flight (if you attempt to enqueue an additional item, you'd have to wait until there is room in the queue). Use this equation to determine what configuration options you want to use to get the type of throughput you want with the pool.

## Pipelines

You can "link" pools together using the input/output queue instances. For example, if you have two processes that operate on different units of work (or different stages in its consumption), you can use the ouput of one pool as the input to another pool to shuttle the data around. This doesn't do anything special (that you couldn't do yourself), but use of the go-queue interfaces should significantly reduce your boiler plate code.
