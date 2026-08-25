package utils

import (
	"bufio"
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
)

// EventMessage merepresentasikan payload data kejadian real-time yang dikirim via SSE.
type EventMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// SSEBroker mengelola koneksi klien aktif dan mempublikasikan pesan kejadian secara thread-safe.
type SSEBroker struct {
	clients    map[chan string]bool
	newClients chan chan string
	defunct    chan chan string
	messages   chan string
	mu         sync.RWMutex
}

var globalBroker *SSEBroker
var once sync.Once

// GetSSEBroker mengembalikan instansi tunggal (singleton) SSEBroker global.
func GetSSEBroker() *SSEBroker {
	once.Do(func() {
		globalBroker = &SSEBroker{
			clients:    make(map[chan string]bool),
			newClients: make(chan chan string, 10),
			defunct:    make(chan chan string, 10),
			messages:   make(chan string, 100),
		}
		go globalBroker.start()
	})
	return globalBroker
}

func (b *SSEBroker) start() {
	for {
		select {
		case s := <-b.newClients:
			b.mu.Lock()
			b.clients[s] = true
			b.mu.Unlock()
		case s := <-b.defunct:
			b.mu.Lock()
			if _, ok := b.clients[s]; ok {
				delete(b.clients, s)
				close(s)
			}
			b.mu.Unlock()
		case msg := <-b.messages:
			b.mu.RLock()
			for s := range b.clients {
				select {
				case s <- msg:
				default:
					// Klien lambat / buffer penuh, lewati agar tidak memblokir broadcast
				}
			}
			b.mu.RUnlock()
		}
	}
}

// Broadcast mengirimkan pesan kejadian SSE ke seluruh klien terhubung (seperti KDS & WLD).
func (b *SSEBroker) Broadcast(eventType string, data string) {
	formattedMsg := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, data)
	b.messages <- formattedMsg
}

// StreamHandler menyajikan aliran HTTP Server-Sent Events (SSE) berkelanjutan.
func (b *SSEBroker) StreamHandler(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	messageChan := make(chan string, 10)
	b.newClients <- messageChan

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		// Kirim ping pesan selamat datang awal
		fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\"}\n\n")
		_ = w.Flush()

		notify := c.Context().Done()
		for {
			select {
			case <-notify:
				b.defunct <- messageChan
				return
			case msg, ok := <-messageChan:
				if !ok {
					return
				}
				if _, err := fmt.Fprint(w, msg); err != nil {
					b.defunct <- messageChan
					return
				}
				_ = w.Flush()
			}
		}
	})

	return nil
}
