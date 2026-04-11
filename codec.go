package wspulse

import "encoding/json"

// Codec encodes and decodes Frames for transmission over a WebSocket connection.
type Codec interface {
	// Encode serializes f into bytes ready to be sent as a WebSocket frame.
	Encode(f Frame) ([]byte, error)

	// Decode deserializes received WebSocket bytes into a Frame.
	Decode(data []byte) (Frame, error)

	// FrameType returns the WebSocket message type to use when sending.
	// The returned int matches the MessageType constants (MessageText = 1,
	// MessageBinary = 2). Consuming modules cast to MessageType at the
	// adapter boundary.
	FrameType() int
}

// wireFrame is the JSON on-the-wire representation of a Frame.
// Payload is treated as json.RawMessage to avoid base64 encoding.
type wireFrame struct {
	Event   string          `json:"event,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// ---- JSON codec (default) ---------------------------------------------------

type jsonCodec struct{}

// JSONCodec is the default Codec. Frames are encoded as JSON text frames.
// Frame.Payload must be valid JSON bytes (e.g. the output of json.Marshal).
var JSONCodec Codec = jsonCodec{}

func (jsonCodec) FrameType() int { return int(MessageText) }

func (jsonCodec) Encode(f Frame) ([]byte, error) {
	return json.Marshal(wireFrame{
		Event:   f.Event,
		Payload: json.RawMessage(f.Payload),
	})
}

func (jsonCodec) Decode(data []byte) (Frame, error) {
	var wf wireFrame
	if err := json.Unmarshal(data, &wf); err != nil {
		return Frame{}, err
	}
	return Frame{
		Event:   wf.Event,
		Payload: []byte(wf.Payload),
	}, nil
}
