package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/knusbaum/ion"
	"github.com/knusbaum/ion/result"
)

type Pubsub struct {
	mu   sync.RWMutex
	subs []chan string
}

func (p *Pubsub) Shutdown() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, c := range p.subs {
		close(c)
	}
	p.subs = p.subs[:0]
}

func (p *Pubsub) Subscribe() chan string {
	ch := make(chan string)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.subs = append(p.subs, ch)
	return ch
}

func (p *Pubsub) Unsubscribe(ch chan string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	var k int
	for i := 0; i < len(p.subs); i++ {
		if p.subs[i] == ch {
			continue
		}
		p.subs[k] = p.subs[i]
		k++
	}
	p.subs = p.subs[:k]
}

func (p *Pubsub) Publish(s string) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, c := range p.subs {
		select {
		case c <- s:
		default:
		}
	}
}

// listen produces a Seq[Result[net.Conn]] of the connections received
// when listening on tcp!addr.
func listen(ctx context.Context, addr string) ion.Seq[result.Res[net.Conn]] {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		var v *ion.Vec[result.Res[net.Conn]]
		v = v.Append(result.Err[net.Conn](err))
		return v
	}
	log.Printf("Listening on %v", addr)

	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	return ion.StateGen(func() (result.Res[net.Conn], bool) {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Failed to accept connection.\n")
			return result.Err[net.Conn](err), false
		}
		return result.Ok[net.Conn](conn), true
	})
}

// lines reads r and produces a Seq[string] of the text lines contained
// in it.
func lines[T io.Reader](r T) ion.Seq[string] {
	b := bufio.NewReader(r)
	return ion.StateGen(func() (string, bool) {
		s, err := b.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("Error: %v\n", err)
			}
			var ior io.Reader
			ior = r
			if cl, ok := ior.(io.ReadCloser); ok {
				cl.Close()
			}
			return "", false
		}
		return strings.TrimSpace(s), true
	})
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	// Sequence of all connections
	conns := listen(ctx, "localhost:8181")

	// Map a handler over the seq to handle any error Results
	conns = ion.Map(conns, result.Handle[net.Conn](func(err error) {
		log.Printf("Failed to listen: %v", err)
	}))

	// Map publishers over the conns. For each conn, subscribe to the
	// message bus and spawn a goroutine that writes all messages to the conn.
	p := &Pubsub{}
	var wg sync.WaitGroup
	conns = result.Map(conns, func(c net.Conn) net.Conn {
		ch := p.Subscribe()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer p.Unsubscribe(ch)
			for e := range ch {
				_, err := fmt.Fprintf(c, "%s\n", e)
				if err != nil {
					fmt.Printf("ERROR SENDING: %s\n", err)
					return
				}
			}
		}()
		return c
	})

	// Map Seq[Result[net.Conn]] -> Seq[Result[Seq[string]]]
	// For each successful connection, create a sequence of lines produced by the
	// clients on those connections
	ls := result.Map(conns, lines)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt)
		<-c
		cancel()
	}()

	// Iterate over each sequence of lines, starting a goroutine that iterates through
	// those lines, publishing them.
	ls.Iterate(ion.Always(result.Apply(func(e ion.Seq[string]) {
		go e.Iterate(ion.Always(p.Publish))
	})))

	// We reached the end of the conn seq, meaning the listener has closed.
	p.Publish("Server shutting down. Goodbye.")
	p.Shutdown()

	// Wait for all the publishers to finish.
	wg.Wait()
	log.Printf("Shutting down.\n")
}
