package wspulse

import "encoding/json"

// Codec encodes and decodes Messages for transmission over a WebSocket connection.
type Codec interface {
	// Encode serializes m into bytes ready to be sent over a WebSocket connection.
	Encode(m Message) ([]byte, error)

	// Decode deserializes received WebSocket bytes into a Message.
	Decode(data []byte) (Message, error)

	// WireType returns the WebSocket message type to use when sending.
	WireType() MessageType
}

// wireMessage is the JSON on-the-wire representation of a Message.
// Payload is treated as json.RawMessage to avoid base64 encoding.
type wireMessage struct {
	Event   string          `json:"event,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// ---- JSON codec (default) ---------------------------------------------------

type jsonCodec struct{}

// JSONCodec is the default Codec. Messages are encoded as JSON text frames.
// Message.Payload must be valid JSON bytes (e.g. the output of json.Marshal).
var JSONCodec Codec = jsonCodec{}

func (jsonCodec) WireType() MessageType { return TextMessage }

func (jsonCodec) Encode(m Message) ([]byte, error) {
	return json.Marshal(wireMessage{
		Event:   m.Event,
		Payload: json.RawMessage(m.Payload),
	})
}

func (jsonCodec) Decode(data []byte) (Message, error) {
	var wm wireMessage
	if err := json.Unmarshal(data, &wm); err != nil {
		return Message{}, err
	}
	return Message{
		Event:   wm.Event,
		Payload: []byte(wm.Payload),
	}, nil
}
