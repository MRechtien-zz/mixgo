package transport

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

func TestTCPPersistentConnection(t *testing.T) {
	addr, received, accepts := startServer(t)
	sender, err := NewTCP(addr)
	if err != nil {
		t.Fatal(err)
	}
	defer sender.Close()

	if err := sender.Send(context.Background(), []byte("one")); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	expectMessage(t, received, "one")

	if err := sender.Send(context.Background(), []byte("two")); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	expectMessage(t, received, "two")

	if got := <-accepts; got != 1 {
		t.Fatalf("connections = %d, want 1", got)
	}
}

func expectMessage(t *testing.T, received <-chan string, want string) {
	t.Helper()
	select {
	case got := <-received:
		if got != want {
			t.Fatalf("received %q, want %q", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %q", want)
	}
}

func TestTCPClose(t *testing.T) {
	sender, err := NewTCP("127.0.0.1:1")
	if err != nil {
		t.Fatal(err)
	}
	if err := sender.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := sender.Send(context.Background(), []byte("x")); !errors.Is(err, ErrClosed) {
		t.Fatalf("Send() error = %v, want ErrClosed", err)
	}
	if err := sender.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
}

func TestNewTCPRejectsInvalidAddress(t *testing.T) {
	if _, err := NewTCP("not-an-address"); err == nil {
		t.Fatal("NewTCP() succeeded, want error")
	}
}

func TestTCPSendCancelledContext(t *testing.T) {
	sender, err := NewTCP("127.0.0.1:1")
	if err != nil {
		t.Fatal(err)
	}
	defer sender.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sender.Send(ctx, []byte("x")); err == nil {
		t.Fatal("Send() succeeded, want error")
	}
}

func startServer(t *testing.T) (addr string, messages <-chan string, accepts <-chan int) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	msgCh := make(chan string, 16)
	acceptCh := make(chan int, 16)
	go func() {
		connections := 0
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			connections++
			acceptCh <- connections
			go func(conn net.Conn) {
				defer conn.Close()
				for {
					buf := make([]byte, 64)
					n, err := conn.Read(buf)
					if err != nil {
						if err != io.EOF {
							return
						}
						return
					}
					msgCh <- string(buf[:n])
				}
			}(conn)
		}
	}()
	return listener.Addr().String(), msgCh, acceptCh
}
