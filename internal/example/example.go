package example

func WorkerFxHelloWorld(workerID int, input interface{}) (output interface{}) {
	return "Hello, World!"
}

func WorkerFxIncrement(workerID int, input interface{}) (output interface{}) {
	switch v := input.(type) {
	case int:
		return v + 1
	}
	return nil
}

func WorkerFxPanic(workerID int, input interface{}) (output interface{}) {
	panic("Look ma, I panicked")
}
