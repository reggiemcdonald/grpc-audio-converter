package converterservice_test

import (
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand"
	"testing"
	"time"
)

type mockJob struct {
	done *bool
	seconds int
}

func (m *mockJob) Start() {
	log.Print("starting job...")
	time.Sleep(time.Duration(m.seconds) * time.Second)
	*m.done = true
	log.Printf("complete job.")
}

func newTestJobQueueConfig() *converterservice.JobQueueConfiguration {
	return &converterservice.JobQueueConfiguration{
		Concurrency: 5,
		QueueSize: 100,
	}
}

func newMockJob(done *bool, seconds int) converterservice.FileConverterJob {
	return &mockJob{
		done: done,
		seconds: seconds,
	}
}

func check(bools []bool, allDone chan bool) {
	for {
		done := true
		for _, b := range bools {
			done = done && b
		}
		if done {
			allDone <- true
			return
		}
	}
}

func TestNewJobQueue(t *testing.T) {
	config := newTestJobQueueConfig()
	queue := converterservice.NewJobQueue(config)
	assert.NotNil(t, queue)

	config.Concurrency = 10
	config.QueueSize = 1000
	queue = converterservice.NewJobQueue(config)
	assert.NotNil(t, queue)
}

func TestJobQueue_Start(t *testing.T) {
	config := newTestJobQueueConfig()
	queue := converterservice.NewJobQueue(config)
	err := queue.Start()
	assert.Nil(t, err)
	assert.True(t, queue.Running())
	queue.Stop()
}

func TestJobQueue_Stop(t *testing.T) {
	config := newTestJobQueueConfig()
	queue := converterservice.NewJobQueue(config)
	err := queue.Start()
	assert.Nil(t, err)
	assert.True(t, queue.Running())
	queue.Stop()
	assert.False(t, queue.Running())
}

func TestJobQueue_Enqueue_PoolSize(t *testing.T) {
	config := newTestJobQueueConfig()
	queue := converterservice.NewJobQueue(config)
	defer queue.Stop()
	if err := queue.Start(); err != nil {
		t.Fatal("failed to start queue")
	}
	bools := []bool{false, false, false, false, false}
	allDone := make(chan bool)
	for i, _ := range bools {
		queue.Enqueue(newMockJob(&bools[i], 5))
	}
	go check(bools, allDone)
	timeout := time.After(6 * time.Second)
	select {
	case <- allDone:
	case <- timeout:
		t.Fatal("timeout waiting for jobs to finish")
	}
}

func TestJobQueue_Enqueue_MoreJobs(t *testing.T) {
	config := newTestJobQueueConfig()
	queue := converterservice.NewJobQueue(config)
	defer queue.Stop()
	if err := queue.Start(); err != nil {
		t.Fatal("failed to start queue")
	}
	bools := make([]bool, 10)
	for i, _ := range bools {
		bools[i] = false
	}
	allDone := make(chan bool)
	for i, _ := range bools {
		queue.Enqueue(newMockJob(&bools[i], 5))
	}
	go check(bools, allDone)
	timeout := time.After(11 * time.Second)
	select {
	case <- allDone:
	case <- timeout:
		t.Fatal("timeout waiting for jobs to finish")
	}
}

func TestJobQueue_Enqueue_StagnatedCompletion(t *testing.T) {
	config := newTestJobQueueConfig()
	queue := converterservice.NewJobQueue(config)
	defer queue.Stop()
	if err := queue.Start(); err != nil {
		t.Fatal("failed to start queue")
	}
	bools := make([]bool, 10)
	for i, _ := range bools {
		bools[i] = false
	}
	allDone := make(chan bool)
	for i, _ := range bools {
		queue.Enqueue(newMockJob(&bools[i], rand.Intn(5)))
	}
	go check(bools, allDone)
	timeout := time.After(time.Duration(5 * len(bools)) * time.Second)
	select {
	case <- allDone:
	case <- timeout:
		t.Fatal("timeout waiting for jobs to finish")
	}
}

func TestJobQueue_NotOverConcurrency(t *testing.T) {
	config := newTestJobQueueConfig()
	queue := converterservice.NewJobQueue(config)
	defer queue.Stop()
	if err := queue.Start(); err != nil {
		t.Fatal("failed to start queue")
	}
	bools := make([]bool, 10)
	for i, _ := range bools {
		bools[i] = false
	}
	for i, _ := range bools {
		queue.Enqueue(newMockJob(&bools[i], 5))
	}
	<- time.After(6 * time.Second)
	count := 0
	for _, b := range bools {
		if b {
			count += 1
		}
	}
	assert.Equal(t, 5, count)
}
