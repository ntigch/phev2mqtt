package cmd

import (
	"testing"
	"time"

	"github.com/ntigch/phev2mqtt/client"
)

func TestEnqueueDoesNotBlockWhenFull(t *testing.T) {
	m := &mqttClient{commands: make(chan carCommand, commandQueueSize)}
	noop := func(*client.Client) error { return nil }

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < commandQueueSize+5; i++ {
			m.enqueue("test", noop)
		}
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("enqueue blocked on a full queue")
	}
	if got := len(m.commands); got != commandQueueSize {
		t.Errorf("queue length = %d, want %d", got, commandQueueSize)
	}
}

func TestRunCommandsInOrder(t *testing.T) {
	m := &mqttClient{commands: make(chan carCommand, commandQueueSize)}
	phev, err := client.New()
	if err != nil {
		t.Fatal(err)
	}
	m.setPhev(phev)

	got := make(chan int, 3)
	for i := 0; i < 3; i++ {
		m.enqueue("test", func(c *client.Client) error {
			if c != phev {
				t.Errorf("command got client %p, want %p", c, phev)
			}
			got <- i
			return nil
		})
	}
	close(m.commands)
	m.runCommands()
	close(got)
	want := 0
	for i := range got {
		if i != want {
			t.Errorf("command %d ran in position %d", i, want)
		}
		want++
	}
	if want != 3 {
		t.Errorf("ran %d commands, want 3", want)
	}
}
