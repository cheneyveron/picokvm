package kvm

import (
	"encoding/json"
	"sync"

	"github.com/pion/webrtc/v4"
)

// HID reports must reach the USB gadget in the order the browser sent them.
// Dispatching every RPC message in its own goroutine lets a key release be
// written before its press, which leaves the key held on the target and the
// host (most visibly a BIOS) auto-repeats it. Keyboard and mouse reports are
// therefore handled by one ordered worker each; mouse gets its own queue so a
// host that never polls the mouse endpoints (BIOS) cannot stall the keyboard.
var keyboardRPCMethods = map[string]bool{
	"keyboardReport":    true,
	"keypressReport":    true,
	"keypressKeepAlive": true,
}

var mouseRPCMethods = map[string]bool{
	"absMouseReport": true,
	"relMouseReport": true,
	"wheelReport":    true,
}

const hidRPCQueueLen = 256

type rpcDispatcher struct {
	keyboard chan webrtc.DataChannelMessage
	mouse    chan webrtc.DataChannelMessage
	done     chan struct{}
	once     sync.Once
}

func newRPCDispatcher(session *Session) *rpcDispatcher {
	d := &rpcDispatcher{
		keyboard: make(chan webrtc.DataChannelMessage, hidRPCQueueLen),
		mouse:    make(chan webrtc.DataChannelMessage, hidRPCQueueLen),
		done:     make(chan struct{}),
	}
	go d.run(d.keyboard, session)
	go d.run(d.mouse, session)
	return d
}

func (d *rpcDispatcher) run(queue chan webrtc.DataChannelMessage, session *Session) {
	for {
		select {
		case msg := <-queue:
			onRPCMessage(msg, session)
		case <-d.done:
			return
		}
	}
}

func (d *rpcDispatcher) dispatch(msg webrtc.DataChannelMessage, session *Session) {
	var head struct {
		Method string `json:"method"`
	}
	if err := json.Unmarshal(msg.Data, &head); err == nil {
		switch {
		case keyboardRPCMethods[head.Method]:
			select { // blocking send keeps order even when the queue is full
			case d.keyboard <- msg:
			case <-d.done:
			}
			return
		case mouseRPCMethods[head.Method]:
			select {
			case d.mouse <- msg:
			default: // drop stale mouse motion rather than block the data channel
			}
			return
		}
	}
	go onRPCMessage(msg, session)
}

// close stops the workers; the queues are never closed, so a late OnMessage
// cannot panic on a send to a closed channel.
func (d *rpcDispatcher) close() {
	d.once.Do(func() { close(d.done) })
}
