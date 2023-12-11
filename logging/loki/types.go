package loki

import (
	"encoding/json"
	"fmt"
	"sort"
	strings "strings"
	"time"

	"github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/common/model"
)

// EntryWithLabels is a log entry with labels.
type EntryWithLabels struct {
	Labels model.LabelSet
	logproto.Entry
}

// StreamWithLables is a log stream with labels.
type StreamWithLables struct {
	LabelSet model.LabelSet
	logproto.Stream
}

func getStreamWithLables(stream logproto.Stream) StreamWithLables {
	return StreamWithLables{
		Stream:   stream,
		LabelSet: stringToLabelsSet(stream.Labels),
	}
}

func labelsMapToString(ls model.LabelSet, without ...model.LabelName) string {
	lstrs := make([]string, 0, len(ls))
Outer:
	for l, v := range ls {
		for _, w := range without {
			if l == w {
				continue Outer
			}
		}
		lstrs = append(lstrs, fmt.Sprintf("%s=%q", l, v))
	}

	sort.Strings(lstrs)
	return fmt.Sprintf("{%s}", strings.Join(lstrs, ", "))
}

func stringToLabelsSet(str string) model.LabelSet {
	m := map[model.LabelName]model.LabelValue{}
	if len(str) < 2 {
		return m
	}
	strs := strings.Split(str[1:len(str)-1], ",") // k="v"
	for _, v := range strs {
		vv := strings.Split(v, "=")
		if len(vv) == 2 && len(vv[1]) >= 2 { //[]{k,v}
			m[model.LabelName(strings.TrimSpace(vv[0]))] = model.LabelValue(strings.TrimSpace(vv[1][1 : len(vv[1])-1]))
		}
	}
	return m
}

// batch holds pending log streams waiting to be sent to Loki, and it's used
// to reduce the number of push requests to Loki aggregating multiple log streams
// and entries in a single batch request. In case of multi-tenant Promtail, log
// streams for each tenant are stored in a dedicated batch.
type batch struct {
	Streams   map[string]*StreamWithLables `json:"streams,omitempty"`
	bytes     int
	createdAt time.Time
}

func newBatch(entries ...EntryWithLabels) *batch {
	b := &batch{
		Streams:   map[string]*StreamWithLables{},
		bytes:     0,
		createdAt: time.Now(),
	}

	// Add entries to the batch
	for _, entry := range entries {
		b.add(entry)
	}

	return b
}

// add an entry to the batch
func (b *batch) add(entry EntryWithLabels) {
	b.bytes += len(entry.Line)

	// Append the entry to an already existing stream (if any)
	labels := labelsMapToString(entry.Labels)
	if stream, ok := b.Streams[labels]; ok {
		stream.Entries = append(stream.Entries, entry.Entry)
		return
	}

	// Add the entry as a new stream
	b.Streams[labels] = &StreamWithLables{
		LabelSet: entry.Labels,
		Stream: logproto.Stream{
			Labels:  labels,
			Entries: []logproto.Entry{entry.Entry},
		},
	}
}

// sizeBytes returns the current batch size in bytes
// func (b *batch) sizeBytes() int {
//	return b.bytes
//}

// sizeBytesAfter returns the size of the batch after the input entry
// will be added to the batch itself
func (b *batch) sizeBytesAfter(entry logproto.Entry) int {
	return b.bytes + len(entry.Line)
}

// age of the batch since its creation
func (b *batch) age() time.Duration {
	return time.Since(b.createdAt)
}

// creates push request and returns it, together with number of entries
func (b *batch) createPushRequest() (*logproto.PushRequest, int) {
	req := logproto.PushRequest{
		Streams: make([]logproto.Stream, 0, len(b.Streams)),
	}

	entriesCount := 0
	for _, stream := range b.Streams {
		req.Streams = append(req.Streams, stream.Stream)
		entriesCount += len(stream.Entries)
	}
	return &req, entriesCount
}

/*
// encode the batch as snappy-compressed push request, and returns
// the encoded bytes and the number of encoded entries
func (b *batch) encode() ([]byte, int, error) {
	req, entriesCount := b.createPushRequest()
	buf, err := proto.Marshal(req)
	if err != nil {
		return nil, 0, err
	}
	buf = snappy.Encode(nil, buf)
	return buf, entriesCount, nil
}
*/

// JSONRequest is a json request for loki
type JSONRequest struct {
	Streams []JSONRequestStream `json:"streams,omitempty"`
}

// JSONRequestStream contains the log within a JSONRequest
type JSONRequestStream struct {
	Stream model.LabelSet         `json:"stream,omitempty"`
	Values []JSONRequestValueItem `json:"values,omitempty"`
}

// JSONRequestValueItem is an array of string
type JSONRequestValueItem []string

// EncodeJSON encode a batch to a JSON object
func (b *batch) EncodeJSON() ([]byte, error) {
	streams := []JSONRequestStream{}
	for _, stream := range b.Streams {
		values := []JSONRequestValueItem{}
		for _, entry := range stream.Entries {
			values = append(values, []string{
				fmt.Sprintf("%d", entry.Timestamp.UnixNano()),
				entry.Line,
			})
		}
		jrs := JSONRequestStream{
			Stream: stream.LabelSet,
			Values: values,
		}
		streams = append(streams, jrs)
	}
	return json.Marshal(&JSONRequest{
		Streams: streams,
	})
}
