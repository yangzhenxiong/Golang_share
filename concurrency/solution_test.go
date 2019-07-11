package solution

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

var (
	// 10秒内 总共完成的任务数
	TOTAL = 0
	// A buffered channel that we can send work requests on.
	JobQueue chan Job
)

type Payload struct {
	name string
}

func (p *Payload) Handler() {
	result := Fibo(40)
	fmt.Println(p.name+" finished calculating, the 40th num is ", strconv.Itoa(result))
	TOTAL++
}

// Job represents the job to be run
type Job struct {
	Payload Payload
}

// Worker represents the worker that executes the job
type Worker struct {
	WorkerPool chan chan Job
	JobChannel chan Job
	quit       chan bool
	name       string
}

func NewWorker(workerPool chan chan Job, name string) Worker {
	return Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool),
		name:       name}
}

// Start method starts the run loop for the worker, listening for a quit channel in
// case we need to stop it
func (w Worker) Start() {
	go func() {
		for {
			// register the current worker into the worker queue.
			// fmt.Println("worker name: ", w.name, " has got the resource.")
			w.WorkerPool <- w.JobChannel
			select {
			case job := <-w.JobChannel:
				// we have received a work request.
				// fmt.Println("Receive a Job: " + job.Payload.name)
				job.Payload.Handler()

			case <-w.quit:
				// we have received a signal to stop
				return
			}
		}
	}()
}

// Stop signals the worker to stop listening for work requests.
func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

type Dispatcher struct {
	// A pool of workers channels that are registered with the dispatcher
	WorkerPool chan chan Job
	maxWorkers int
}

func NewDispatcher(maxWorkers int) *Dispatcher {
	pool := make(chan chan Job, maxWorkers)
	return &Dispatcher{WorkerPool: pool, maxWorkers: maxWorkers}
}

func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.WorkerPool, strconv.Itoa(i))
		worker.Start()
	}
	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-JobQueue:
			// a job request has been received
			// fmt.Println("a job has been received: " + job.Payload.name)
			go func(job Job) {
				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				jobChannel := <-d.WorkerPool
				// dispatch the job to the worker job channel
				jobChannel <- job
			}(job)
		}
	}
}

//并发：计算第40个斐波那契数字，10秒内能计算几次
func TestConCurrencyMethod(t *testing.T) {
	JobQueue = make(chan Job, 10)
	dispatcher := NewDispatcher(10)
	dispatcher.Run()
	for i := 0; i < 1000; i++ {
		index := strconv.Itoa(i)
		payload := Payload{"Allen" + index}
		work := Job{Payload: payload}
		JobQueue <- work
	}
	time.Sleep(time.Second * 10)
	fmt.Println("EDN, Total is: ", TOTAL)
}

//串行：计算第40个斐波那契数字，10秒内能计算几次
func TestSerialMethod(t *testing.T) {
	begin := time.Now().Unix()
	for i := 0; i < 1000; i++ {
		result := Fibo(40)
		fmt.Println(strconv.Itoa(i)+" finished calculating, the 40th num is ", strconv.Itoa(result))
		now := time.Now().Unix()
		cost := now - begin
		if cost > 10 {
			fmt.Println("END, Total is:", i+1)
			break
		}
	}
}

// Worker的任务，计算斐波那契数列的第N个数
func Fibo(n int) int {
	if n == 0 {
		return 0
	} else if n == 1 {
		return 1
	} else if n > 1 {
		return Fibo(n-1) + Fibo(n-2)
	} else {
		return -1
	}
}

func TestFibo(t *testing.T) {
	begin := time.Now().Unix()
	result := Fibo(40)
	end := time.Now().Unix()
	fmt.Println("Finish Fibo 40: ", result, (end - begin))
}
