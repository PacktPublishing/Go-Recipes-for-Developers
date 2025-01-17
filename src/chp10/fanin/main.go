package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type InputPayload struct {
	Id int
	// Add payload data fields here
}

type Stage2Payload struct {
	Id int
	// Stage2 data fields here
}

type Stage3Payload struct {
	Id int
	// Stage3 data fields here
}

type OutputPayload struct {
	Id int
}

func processData(id int) error {
	// Process data
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
	// There may be an error
	if rand.Intn(100) < 10 {
		return fmt.Errorf("Processing failure for id: %d", id)
	}
	return nil
}

type PipelineError struct {
	Stage   int
	Payload any
	Err     error
}

func (p PipelineError) Error() string {
	return fmt.Sprintf("Pipeline error at stage: %d. Payload: %v. Cause: %s", p.Stage, p.Payload, p.Err)
}

func Stage1(input <-chan InputPayload, errCh chan<- error) <-chan Stage2Payload {
	output := make(chan Stage2Payload)
	go func() {
		defer close(output)
		// Process all inputs
		for in := range input {
			// Process data
			err := processData(in.Id)
			if err != nil {
				errCh <- PipelineError{
					Stage:   1,
					Payload: in,
					Err:     err,
				}
				continue
			}
			output <- Stage2Payload{
				Id: in.Id,
			}
		}
	}()
	return output
}

func Stage2(input <-chan Stage2Payload, errCh chan<- error) <-chan Stage3Payload {
	output := make(chan Stage3Payload)
	go func() {
		defer close(output)
		// Process all inputs
		for in := range input {
			// Process data
			err := processData(in.Id)
			if err != nil {
				errCh <- PipelineError{
					Stage:   2,
					Payload: in,
					Err:     err,
				}
				continue
			}
			output <- Stage3Payload{
				Id: in.Id,
			}
		}
	}()
	return output
}

func Stage3(input <-chan Stage3Payload, errCh chan<- error) <-chan OutputPayload {
	output := make(chan OutputPayload)
	go func() {
		defer close(output)
		// Process all inputs
		for in := range input {
			// Process data
			err := processData(in.Id)
			if err != nil {
				errCh <- PipelineError{
					Stage:   3,
					Payload: in,
					Err:     err,
				}
				continue
			}
			output <- OutputPayload{
				Id: in.Id,
			}
		}
	}()
	return output
}

func fanIn(inputs []<-chan OutputPayload) <-chan OutputPayload {

	result := make(chan OutputPayload)

	// Listen to input channels in separate goroutines
	inputWg := sync.WaitGroup{}
	for inputIndex := range inputs {
		inputWg.Add(1)
		go func(index int) {
			defer inputWg.Done()
			for data := range inputs[index] {
				// Send the data to the output
				result <- data
			}
		}(inputIndex)
	}

	// When all input channels are closed, close the fan in ch
	go func() {
		inputWg.Wait()
		close(result)
	}()

	return result
}

func main() {
	errCh := make(chan error)
	inputCh := make(chan InputPayload)

	poolSize := 5
	outputs := make([]<-chan OutputPayload, 0)
	// All Stage1 goroutines listen to a single input channel
	for i := 0; i < poolSize; i++ {
		outputCh1 := Stage1(inputCh, errCh)
		outputCh2 := Stage2(outputCh1, errCh)
		outputCh3 := Stage3(outputCh2, errCh)
		outputs = append(outputs, outputCh3)
	}

	outputCh := fanIn(outputs)

	// Feed input asynchronously
	go func() {
		defer close(inputCh)
		for i := 0; i < 1000; i++ {
			inputCh <- InputPayload{
				Id: i,
			}
		}
	}()

	// Listen to the error channel asynchronously
	go func() {
		for err := range errCh {
			fmt.Println(err)
		}
	}()

	// Read outputs
	for out := range outputCh {
		fmt.Println(out)
	}
	// Close the error channel
	close(errCh)

}
