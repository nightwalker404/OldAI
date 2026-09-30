package proto

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

const MaxLine = 64 << 10

type Message struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

func Write(w io.Writer, m Message) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	_, err = w.Write(append(b, '\n'))

	return err
}

type Reader struct{ sc *bufio.Scanner }

func NewReader(r io.Reader) *Reader {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 4096), MaxLine)
	return &Reader{sc: sc}
}

func (r *Reader) Read() (Message, error) {
	if !r.sc.Scan() {
		if err := r.sc.Err(); err != nil {
			return Message{}, err
		}
		return Message{}, io.EOF
	}
	var m Message
	if err := json.Unmarshal(r.sc.Bytes(), &m); err != nil {
		return Message{}, fmt.Errorf("malformed message: %w", err)
	}
	return m, nil
}
