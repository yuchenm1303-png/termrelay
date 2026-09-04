package chatgptweb

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

type StreamEventType string

const (
	StreamEventUpdate StreamEventType = "update"
	StreamEventDone   StreamEventType = "done"
)

type StreamEvent struct {
	Type            StreamEventType
	Snapshot        Snapshot
	FinalDelta      string
	CommentaryDelta string
}

type StreamParser struct {
	scanner      *bufio.Scanner
	accumulator  *PatchAccumulator
	pendingData  []string
	lastSnapshot Snapshot
	done         bool
}

func NewStreamParser(r io.Reader) *StreamParser {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 8*1024*1024)
	return &StreamParser{scanner: scanner, accumulator: NewPatchAccumulator()}
}

func (p *StreamParser) Next() (StreamEvent, error) {
	if p == nil || p.scanner == nil {
		return StreamEvent{}, errors.New("chatgptweb: nil stream parser")
	}
	if p.done {
		return StreamEvent{}, io.EOF
	}
	for p.scanner.Scan() {
		line := p.scanner.Text()
		if line == "" {
			if len(p.pendingData) == 0 {
				continue
			}
			return p.consumePending()
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		if strings.HasPrefix(line, "data:") {
			p.pendingData = append(p.pendingData, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
	if err := p.scanner.Err(); err != nil {
		return StreamEvent{}, fmt.Errorf("chatgptweb: read SSE: %w", err)
	}
	if len(p.pendingData) > 0 {
		return p.consumePending()
	}
	return StreamEvent{}, io.EOF
}

func (p *StreamParser) consumePending() (StreamEvent, error) {
	data := strings.Join(p.pendingData, "\n")
	p.pendingData = p.pendingData[:0]
	if data == "[DONE]" {
		p.done = true
		snapshot := p.accumulator.Snapshot()
		return StreamEvent{Type: StreamEventDone, Snapshot: snapshot}, nil
	}
	if data == "" {
		return p.Next()
	}
	snapshot, err := p.accumulator.ApplyJSON([]byte(data))
	if err != nil {
		return StreamEvent{}, err
	}
	event := StreamEvent{
		Type:            StreamEventUpdate,
		Snapshot:        snapshot,
		FinalDelta:      suffixDelta(p.lastSnapshot.FinalText, snapshot.FinalText),
		CommentaryDelta: suffixDelta(p.lastSnapshot.CommentaryText, snapshot.CommentaryText),
	}
	p.lastSnapshot = snapshot
	return event, nil
}

func suffixDelta(previous, current string) string {
	if previous == "" {
		return current
	}
	if strings.HasPrefix(current, previous) {
		return current[len(previous):]
	}
	if current == previous {
		return ""
	}
	return current
}
