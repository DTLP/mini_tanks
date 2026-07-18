package game

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/DTLP/mini_tanks/internal/actors"
	"github.com/DTLP/mini_tanks/internal/levels"
)

// TankSnap is a JSON-safe projection of an actors.Tank for network transport.
type TankSnap struct {
	Name            string
	Player          int
	IsPlayer        bool
	X, Y            float64
	HullAngle       float64
	TurretAngle     float64
	HullImage       string
	TurretImage     string
	Width           float64
	Height          float64
	TurretWidth     float64
	TurretHeight    float64
	TurretLength    float64
	Health          int
	MaxHealth       int
	HealthBarWidth  int
	HealthBarHeight int
	ReloadBarWidth  int
	ReloadBarHeight int
	ReloadTimer     float64
	ReloadTime      float64
	Projectiles     []actors.Projectile
	Explosions      []actors.Explosion
	LastDamagedBy   string
}

// BlockSnap is a JSON-safe projection of a levels.LevelBlock for network
// transport. It carries the collision walls so the client can render raycast
// shadows identically, plus the texture crop used by scene.drawLevelObjects.
type BlockSnap struct {
	X, Y, Width, Height float64
	Health              int
	Base                bool
	Border              bool
	Destructible        bool
	Collidable          bool
	ImagePath           string
	ImageX, ImageY      int
	ImageWidth, ImageHeight int
	Blocks              []levels.Block
}

// Snapshot is a full server-authoritative world state sent to clients.
type Snapshot struct {
	LevelNum int
	GameOver bool
	Tanks    []TankSnap
	Blocks   []BlockSnap
	KillFeed []actors.KillFeedEntry
}

// Msg is a wire message. Type is one of "input", "snapshot", "hello", "bye".
type Msg struct {
	Type     string        `json:"t"`
	Input    *actors.Input `json:"i,omitempty"`
	Snapshot *Snapshot     `json:"s,omitempty"`
}

// NetConn is the network connection handle. Only the writer goroutine writes
// to conn; only the reader goroutine reads from conn. The main loop interacts
// with the connection via methods.
type NetConn struct {
	conn net.Conn

	sendCh     chan *Msg // writer drains this
	writerDone chan struct{}

	mu          sync.Mutex
	latestInput *actors.Input
	latestSnap  *Snapshot

	done chan struct{}
	once sync.Once
	err  atomic.Value // stores error

	listener net.Listener
	ready    chan struct{}
}

// newNetConn returns a NetConn with its channels initialized and no peer yet.
func newNetConn() *NetConn {
	return &NetConn{
		sendCh:     make(chan *Msg, 8),
		writerDone: make(chan struct{}),
		done:       make(chan struct{}),
		ready:      make(chan struct{}),
	}
}

// closeDone closes done exactly once. It is safe to call from any goroutine.
func (c *NetConn) closeDone() {
	c.once.Do(func() { close(c.done) })
}

// Send queues a message for the writer goroutine. It never blocks the game
// loop: if the writer is busy or the queue is full, the message is dropped.
func (c *NetConn) Send(m *Msg) {
	select {
	case c.sendCh <- m:
	default:
		// Drop if the writer is busy / the channel is full.
	}
}

// GetInput returns the latest received remote Input and clears it. Returns nil
// if none.
func (c *NetConn) GetInput() *actors.Input {
	c.mu.Lock()
	defer c.mu.Unlock()
	in := c.latestInput
	c.latestInput = nil
	return in
}

// PeekInput returns the latest received remote Input WITHOUT clearing. For
// sticky apply.
func (c *NetConn) PeekInput() *actors.Input {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.latestInput
}

// SetInput stores the latest remote Input, overwriting any previous one.
func (c *NetConn) SetInput(in *actors.Input) {
	c.mu.Lock()
	c.latestInput = in
	c.mu.Unlock()
}

// GetSnapshot returns the latest received Snapshot and clears it. Returns nil
// if none.
func (c *NetConn) GetSnapshot() *Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	s := c.latestSnap
	c.latestSnap = nil
	return s
}

// SetSnapshot stores the latest Snapshot, overwriting any previous one.
func (c *NetConn) SetSnapshot(s *Snapshot) {
	c.mu.Lock()
	c.latestSnap = s
	c.mu.Unlock()
}

// Done fires when the connection has terminated for any reason.
func (c *NetConn) Done() <-chan struct{} { return c.done }

// Err returns the error that terminated the connection, or nil if the
// connection is still alive or was closed without an error.
func (c *NetConn) Err() error {
	if v := c.err.Load(); v != nil {
		return v.(error)
	}
	return nil
}

// Close shuts the connection down. It is idempotent and safe to call from any
// goroutine.
func (c *NetConn) Close() {
	c.closeDone()
	if c.conn != nil {
		_ = c.conn.Close()
	}
	if c.listener != nil {
		_ = c.listener.Close()
	}
}

// Ready fires once the connection has a peer and the reader/writer goroutines
// are running.
func (c *NetConn) Ready() <-chan struct{} { return c.ready }

// IsConnected reports whether the connection has completed its handshake.
func (c *NetConn) IsConnected() bool {
	select {
	case <-c.ready:
		return true
	default:
		return false
	}
}

// launch wires the underlying connection and starts one persistent reader
// goroutine and one persistent writer goroutine. It is called exactly once per
// NetConn: synchronously from Join, and from the accept goroutine on the Host.
func (c *NetConn) launch(nc net.Conn) {
	c.conn = nc

	// Writer goroutine: owns all writes to conn.
	go func() {
		defer close(c.writerDone)
		bw := bufio.NewWriter(nc)
		for {
			select {
			case <-c.done:
				return
			case m := <-c.sendCh:
				if err := writeMsg(bw, m); err != nil {
					c.err.Store(err)
					c.closeDone()
					return
				}
				if err := bw.Flush(); err != nil {
					c.err.Store(err)
					c.closeDone()
					return
				}
			}
		}
	}()

	// Reader goroutine: owns all reads from conn.
	go func() {
		br := bufio.NewReader(nc)
		for {
			m, err := readMsg(br)
			if err != nil {
				c.err.Store(err)
				c.closeDone()
				return
			}
			switch m.Type {
			case "input":
				c.SetInput(m.Input) // Keep the LATEST input; overwrite.
			case "snapshot":
				if m.Snapshot != nil {
					c.SetSnapshot(m.Snapshot) // Keep the LATEST snapshot.
				}
				// "hello" and "bye" are ignored for now.
			}
		}
	}()

	// Signal readiness once the reader/writer goroutines are running.
	select {
	case <-c.ready:
	default:
		close(c.ready)
	}
}

// Host starts listening on listenAddr and accepts ONE client. It returns a
// *NetConn immediately; the connection completes asynchronously. Wait on
// Ready() before expecting snapshots or inputs.
func Host(listenAddr string) (*NetConn, error) {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	conn := newNetConn()
	conn.listener = ln

	// Accept goroutine: accept ONE client, then start reader/writer.
	go func() {
		nc, err := ln.Accept()
		if err != nil {
			conn.err.Store(err)
			conn.closeDone()
			return
		}
		conn.launch(nc)
	}()

	return conn, nil
}

// Join connects to a remote host and returns a ready-to-use *NetConn. The
// reader/writer goroutines start immediately, so Ready() is already closed.
func Join(remoteAddr string) (*NetConn, error) {
	nc, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		return nil, err
	}
	conn := newNetConn()
	conn.launch(nc)
	// Let the host know we are here; the host's Accept already completed.
	conn.Send(&Msg{Type: "hello"})
	return conn, nil
}

// SendInput queues a remote-input message to the peer.
func (c *NetConn) SendInput(in actors.Input) {
	c.Send(&Msg{Type: "input", Input: &in})
}

// SendSnapshot queues a world snapshot message to the peer.
func (c *NetConn) SendSnapshot(s Snapshot) {
	c.Send(&Msg{Type: "snapshot", Snapshot: &s})
}

// writeMsg marshals m and writes it as a length-prefixed frame: a 4-byte
// big-endian length followed by the JSON payload.
func writeMsg(w io.Writer, m *Msg) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(b)))
	if _, err := w.Write(hdr[:]); err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

// readMsg reads one length-prefixed frame from r and unmarshals it.
func readMsg(r *bufio.Reader) (*Msg, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return nil, err
	}
	n := binary.BigEndian.Uint32(hdr[:])
	if n > 32<<20 { // 32 MiB safety cap
		return nil, errors.New("netplay: message too large")
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	var m Msg
	if err := json.Unmarshal(buf, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
