package loki

import (
	"context"
	"fmt"
	"time"

	"github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/common/model"
	"go.uber.org/zap/zapcore"

	"github.com/kiga-hub/arc/logging/conf"
)

const (
	clientBatchSize = 1024
	clientBatchWait = 200 * time.Millisecond
)

// NewLokiCore create a new LokiCore
func NewLokiCore(logConfig *conf.LogConfig) (zapcore.Core, error) {
	lokiClient, err := NewLokiClient(logConfig.LokiAddr)
	if err != nil {
		return nil, err
	}
	c := &Core{
		SendLevel:     logLevelToZapLevel(logConfig.Level),
		ContextFields: []zapcore.Field{},
		URL:           logConfig.LokiAddr,
		client:        lokiClient,
		entryChan:     make(chan EntryWithLabels),
	}
	go c.run()
	return c, nil
}

func logLevelToZapLevel(level string) zapcore.Level {
	switch level {
	case "PANIC":
		return zapcore.PanicLevel
	case "DEBUG":
		return zapcore.DebugLevel
	case "ERROR":
		return zapcore.ErrorLevel
	case "WARN":
		return zapcore.WarnLevel
	default:
		return zapcore.InfoLevel
	}
}

// Core is a zap core for logging to loki
type Core struct {
	SendLevel     zapcore.Level // default: zapCore.InfoLevel
	ContextFields []zapcore.Field
	URL           string
	client        *Client
	entryChan     chan EntryWithLabels
}

// Enabled returns true if the given level is at or above this level.
func (c *Core) Enabled(level zapcore.Level) bool {
	return c.SendLevel.Enabled(level)
}

// With adds structured context to the Core.
func (c *Core) With(fs []zapcore.Field) zapcore.Core {
	c.ContextFields = fs
	return c
}

// Check determines whether the supplied Entry should be logged (using the
// embedded LevelEnabler and possibly some extra logic). If the entry
// should be logged, the Core adds itself to the CheckedEntry and returns
// the result.
//
// Callers must use Check before calling Write.
func (c *Core) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.SendLevel.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

// Write serializes the Entry and any Fields supplied at the log site and
// writes them to their destination.
//
// If called, Write should always log the Entry and Fields; it should not
// replicate the logic of Check.
func (c *Core) Write(entry zapcore.Entry, fs []zapcore.Field) error {

	enc := zapcore.NewMapObjectEncoder()
	for _, f := range c.ContextFields {
		f.AddTo(enc)
	}
	for _, f := range fs {
		f.AddTo(enc)
	}

	labels := model.LabelSet{}
	for k, v := range enc.Fields {
		//fmt.Println(k)
		//fmt.Println(v)
		labels[model.LabelName(k)] = model.LabelValue(fmt.Sprintf("%v", v))
	}
	labels["level"] = model.LabelValue(entry.Level.String())
	if entry.Stack != "" {
		labels["stack"] = model.LabelValue(entry.Stack)
	}
	l := EntryWithLabels{
		Labels: labels,
		Entry: logproto.Entry{
			Timestamp: time.Now(),
			Line:      entry.Message,
		},
	}
	//fmt.Println(entry.Level)
	//fmt.Println(entry.Message)
	//fmt.Println(entry.Stack)

	c.entryChan <- l
	return nil
}

// Sync flushes buffered logs (if any).
func (c *Core) Sync() error {
	return nil
}

func (c *Core) run() {
	bs := newBatch()

	// Given the client handles multiple batches (1 per tenant) and each batch
	// can be created at a different point in time, we look for batches whose
	// max wait time has been reached every 10 times per BatchWait, so that the
	// maximum delay we have sending batches is 10% of the max waiting time.
	// We apply a cap of 10ms to the ticker, to avoid too frequent checks in
	// case the BatchWait is very low.
	//minWaitCheckFrequency := 10 * time.Millisecond
	maxWaitCheckFrequency := clientBatchWait / 10
	// Condition 'maxWaitCheckFrequency < minWaitCheckFrequency' is always 'false'
	//if maxWaitCheckFrequency < minWaitCheckFrequency {
	//	maxWaitCheckFrequency = minWaitCheckFrequency
	//}

	maxWaitCheck := time.NewTicker(maxWaitCheckFrequency)

	defer func() {
		maxWaitCheck.Stop()
		_, err := c.sendBatch(bs)
		if err != nil {
			fmt.Println(err)
		}
	}()

	for {
		select {
		case e, ok := <-c.entryChan:
			if !ok {
				return
			}
			// If the batch doesn't exist yet, we create a new one with the entry

			// If adding the entry to the batch will increase the size over the max
			// size allowed, we do send the current batch and then create a new one
			if bs.sizeBytesAfter(e.Entry) > clientBatchSize {
				_, err := c.sendBatch(bs)
				if err != nil {
					fmt.Println(err)
				}
				bs = newBatch()
				break
			}

			// The max size of the batch isn't reached, so we can add the entry
			bs.add(e)

		case <-maxWaitCheck.C:
			// Send all batches whose max wait time has been reached
			if bs.bytes <= 0 || bs.age() < clientBatchWait {
				continue
			}

			_, err := c.sendBatch(bs)
			if err != nil {
				fmt.Println(err)
			}
			bs = newBatch()
		}
	}
}

/*
func (c *LokiCore) sendBatchByJSON(batch *batch) (int, error) {
	buf, err := batch.EncodeJSON()
	if err != nil {
		return -1, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	req, err := http.NewRequest("POST", c.URL, bytes.NewReader(buf))
	if err != nil {
		return -1, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", contentTypeJSON)
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		scanner := bufio.NewScanner(io.LimitReader(resp.Body, maxErrMsgLen))
		line := ""
		if scanner.Scan() {
			line = scanner.Text()
		}
		err = fmt.Errorf("server returned HTTP status %s (%d): %s", resp.Status, resp.StatusCode, line)
	}
	return resp.StatusCode, err
}
*/

func (c *Core) sendBatch(batch *batch) (int, error) {
	req, entriesCount := batch.createPushRequest()
	fmt.Printf("sending %d entry within %d bytes...\n", entriesCount, req.Size())
	//fmt.Println(string(buf))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	err := c.client.SendLogs(ctx, req)
	return entriesCount, err
}
