package client

import (
	"net"
	"runtime"
	"testing"
	"time"

	"github.com/buxtronix/phev2mqtt/protocol"
)

// startTestServer returns a listener emulating the car's TCP endpoint,
// accepting a single connection and discarding anything sent to it.
func startTestServer(t *testing.T) (net.Listener, chan net.Conn) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	connCh := make(chan net.Conn, 1)
	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		connCh <- conn
		buf := make([]byte, 1024)
		for {
			if _, err := conn.Read(buf); err != nil {
				return
			}
		}
	}()
	return l, connCh
}

func TestSendsDoNotBlockAfterClose(t *testing.T) {
	l, connCh := startTestServer(t)
	defer l.Close()

	cl, err := New(AddressOption(l.Addr().String()))
	if err != nil {
		t.Fatal(err)
	}
	if err := cl.Connect(); err != nil {
		t.Fatal(err)
	}
	<-connCh
	cl.Close()

	done := make(chan error, 1)
	go func() {
		done <- cl.SetRegister(0x6, []byte{0x3})
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Error("SetRegister after Close: got nil error, want error")
		}
	case <-time.After(3 * time.Second):
		t.Error("SetRegister blocked after Close")
	}

	// Repeated queued sends must return even with the writer gone
	// and the Send buffer full.
	sends := make(chan struct{})
	go func() {
		for i := 0; i < 10; i++ {
			cl.SendMessage(protocol.NewPingRequestMessage(byte(i)))
		}
		close(sends)
	}()
	select {
	case <-sends:
	case <-time.After(3 * time.Second):
		t.Error("SendMessage blocked after Close")
	}
}

func TestListenersClosedOnDisconnect(t *testing.T) {
	l, connCh := startTestServer(t)
	defer l.Close()

	cl, err := New(AddressOption(l.Addr().String()))
	if err != nil {
		t.Fatal(err)
	}
	if err := cl.Connect(); err != nil {
		t.Fatal(err)
	}
	lis := cl.AddListener()

	// Simulate the car dropping the connection.
	conn := <-connCh
	conn.Close()

	for {
		select {
		case _, ok := <-lis.C:
			if !ok {
				return // Channel closed as expected.
			}
		case <-time.After(3 * time.Second):
			t.Fatal("listener channel not closed after disconnect")
		}
	}
}

func TestStartTimeoutThenClose(t *testing.T) {
	old := startTimeout
	startTimeout = 200 * time.Millisecond
	defer func() { startTimeout = old }()

	l, connCh := startTestServer(t)
	defer l.Close()

	cl, err := New(AddressOption(l.Addr().String()))
	if err != nil {
		t.Fatal(err)
	}
	if err := cl.Connect(); err != nil {
		t.Fatal(err)
	}
	conn := <-connCh

	// The car never sends a start request.
	if err := cl.Start(); err == nil {
		t.Fatal("Start: got nil error, want timeout")
	}
	cl.Close()

	// The TCP connection must really be closed, so the car frees its slot.
	// Anything the client had already queued (pings) is read and discarded
	// by the test server, which then sees EOF; wait for that.
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 64)
	for {
		if _, err := conn.Read(buf); err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				t.Fatal("connection still open after Close")
			}
			break
		}
	}
}

func TestGoroutinesExitOnClose(t *testing.T) {
	l, connCh := startTestServer(t)
	defer l.Close()

	before := runtime.NumGoroutine()
	cl, err := New(AddressOption(l.Addr().String()))
	if err != nil {
		t.Fatal(err)
	}
	if err := cl.Connect(); err != nil {
		t.Fatal(err)
	}
	<-connCh
	cl.Close()

	// reader, writer, manage and pinger must all exit. The test server's own
	// goroutines also finish once the connection is closed.
	deadline := time.Now().Add(3 * time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if after := runtime.NumGoroutine(); after > before {
		t.Errorf("goroutines leaked after Close: before=%d after=%d", before, after)
	}
}
