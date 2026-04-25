package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	gosse "github.com/tmaxmax/go-sse"
)

// HeartbeatInterval is the gap between SSE comment heartbeats. We
// pick 25 seconds because Vercel + Cloudflare default idle-timeouts
// are 30-60 seconds; sending a comment every 25s keeps the connection
// open across deployment topologies without spamming the wire.
const HeartbeatInterval = 25 * time.Second

// WriteStream drains events into w using SSE framing. It blocks until
// either events closes (on stream end / cancellation upstream) or the
// HTTP client disconnects.
//
// It owns the `Upgrade` call so callers don't have to know SSE
// header conventions; once Upgrade succeeds, every event written
// flushes immediately. The function is intentionally narrow: it
// doesn't choose a model, build a prompt, or own the Anthropic
// client. Compose with Client.
func WriteStream(w http.ResponseWriter, r *http.Request, events <-chan Event) error {
	session, err := gosse.Upgrade(w, r)
	if err != nil {
		return fmt.Errorf("ai: sse upgrade: %w", err)
	}

	heartbeat := time.NewTicker(HeartbeatInterval)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return nil

		case <-heartbeat.C:
			// SSE comment line keeps proxies quiet without
			// counting as a real event for clients.
			if sendErr := session.Send(&gosse.Message{}); sendErr != nil {
				return nil
			}
			if flushErr := session.Flush(); flushErr != nil {
				return nil
			}

		case ev, ok := <-events:
			if !ok {
				return nil
			}
			if writeErr := writeEvent(session, ev); writeErr != nil {
				// Connection broken; nothing useful to do but
				// stop draining the upstream channel.
				return nil
			}
			// Reset heartbeat on real traffic so we don't double up.
			heartbeat.Reset(HeartbeatInterval)
		}
	}
}

func writeEvent(session *gosse.Session, ev Event) error {
	msg := &gosse.Message{Type: gosse.Type(string(ev.Kind))}
	payload, err := payloadJSON(ev)
	if err != nil {
		return fmt.Errorf("encode %s payload: %w", ev.Kind, err)
	}
	msg.AppendData(string(payload))
	if err := session.Send(msg); err != nil {
		return err
	}
	return session.Flush()
}

// payloadJSON renders the data: line for each event kind. Shapes are
// stable -- clients depend on them; changing a field is a wire break.
func payloadJSON(ev Event) ([]byte, error) {
	switch ev.Kind {
	case EventDelta:
		return json.Marshal(deltaPayload{Text: ev.Delta})
	case EventUsage:
		if ev.Usage == nil {
			return json.Marshal(Usage{})
		}
		return json.Marshal(*ev.Usage)
	case EventError:
		msg := "stream error"
		if ev.Err != nil {
			msg = ev.Err.Error()
		}
		return json.Marshal(errorPayload{Message: msg})
	case EventDone:
		return []byte("{}"), nil
	default:
		return nil, fmt.Errorf("unknown event kind %q", ev.Kind)
	}
}

type deltaPayload struct {
	Text string `json:"text"`
}

type errorPayload struct {
	Message string `json:"message"`
}
