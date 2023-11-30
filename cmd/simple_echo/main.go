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
	"time"

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

// handleConn handles a connection by subscribing to p and spawning a goroutine
// sending any messages from p to the conn while reading from the conn and publishing
// those messages to p.
//
// If cr does not contain a conn, it logs an error and returns p
func handleConn(p *Pubsub, cr result.Res[net.Conn]) *Pubsub {
	if p == nil {
		p = &Pubsub{}
	}
	result.Apply(func(c net.Conn) {
		// Start a goroutine iterating all lines on the connection and publishing them.
		go lines[net.Conn](c).Iterate(ion.Always(p.Publish))
		// Start a goroutine listening for all published messages and write them to the
		// conn.
		go func() {
			ch := p.Subscribe()
			defer p.Unsubscribe(ch)
			for m := range ch {
				_, err := fmt.Fprintf(c, "%s\n", m)
				if err != nil {
					return
				}
			}
		}()
	})(cr)
	return p
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	// Conns is a Seq[Result[net.Conn]]
	conns := listen(ctx, "localhost:8181")

	// Map a handler over the seq to handle any error Results
	conns = ion.Map(conns, result.Handle[net.Conn](func(err error) {
		log.Printf("Failed to listen: %v", err)
	}))

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt)
		<-c
		cancel()
	}()

	// Fold the connections with handleConn, which will handle connections
	// one-by-one, spawning goroutines to do the communications.
	p := ion.Fold(conns, handleConn)

	// We reached the end of the conn seq, meaning the listener has closed.
	p.Publish("Server shutting down. Goodbye.")
	p.Shutdown()

	// Wait for all the publishers to finish.
	// wg.Wait() // No waitgroup available. Need to modify the Fold accumulator
	// Instead, we'll just wait for a second.
	time.Sleep(1 * time.Second)
	log.Printf("Shutting down.\n")
}
