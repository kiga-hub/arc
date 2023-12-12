package wave

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
)

// WriterParam -
type WriterParam struct {
	Out           *os.File
	Channel       int
	SampleRate    int
	BitsPerSample int
}

// Writer -
type Writer struct {
	out            *os.File // 实际写出来的文件和bytes等
	writtenSamples int      // 写入的样本数
	flushCount     int

	riffChunk *RiffChunk
	fmtChunk  *FmtChunk
	dataChunk *DataWriterChunk
}

// New builds a new OGG Opus writer
//
//goland:noinspection GoUnusedExportedFunction
func New(fileName string, sampleRate int, channelCount int) (*Writer, error) {
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}

	param := WriterParam{
		Out:           f,
		Channel:       channelCount,
		SampleRate:    sampleRate,
		BitsPerSample: 16,
	}
	writer, err := NewWriter(param)
	if err != nil {
		return nil, f.Close()
	}
	return writer, nil
}

// NewBuffer -
func NewBuffer(sampleRate int, channelCount int) (*Writer, error) {
	param := WriterParam{
		Channel:       channelCount,
		SampleRate:    sampleRate,
		BitsPerSample: 16,
	}
	writer, err := NewWriter(param)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return writer, nil
}

// NewWriter -
func NewWriter(param WriterParam) (*Writer, error) {
	w := &Writer{}
	w.out = param.Out

	blockSize := uint16(param.BitsPerSample*param.Channel) / 8
	samplesPerSec := uint32(int(blockSize) * param.SampleRate)

	// riff chunk
	w.riffChunk = &RiffChunk{
		ID:         []byte(riffChunkToken),
		FormatType: []byte(waveFormatType),
	}
	// fmt chunk
	w.fmtChunk = &FmtChunk{
		ID:   []byte(fmtChunkToken),
		Size: uint32(fmtChunkSize),
	}
	w.fmtChunk.Data = &WavFmtChunkData{
		WaveFormatType: uint16(1), // PCM
		Channel:        uint16(param.Channel),
		SamplesPerSec:  uint32(param.SampleRate),
		BytesPerSec:    samplesPerSec,
		BlockSize:      blockSize,
		BitsPerSamples: uint16(param.BitsPerSample),
	}
	// data chunk
	w.dataChunk = &DataWriterChunk{
		ID:   []byte(dataChunkToken),
		Data: bytes.NewBuffer([]byte{}),
	}

	return w, nil
}

// WriteSample8 -
func (w *Writer) WriteSample8(samples []uint8) (int, error) {
	buf := new(bytes.Buffer)

	for i := 0; i < len(samples); i++ {
		err := binary.Write(buf, binary.LittleEndian, samples[i])
		if err != nil {
			return 0, err
		}
	}
	n, err := w.Write(buf.Bytes())
	return n, err
}

// WriteSample16 -
func (w *Writer) WriteSample16(samples []int16) (int, error) {
	buf := new(bytes.Buffer)

	for i := 0; i < len(samples); i++ {
		err := binary.Write(buf, binary.LittleEndian, samples[i])
		if err != nil {
			return 0, err
		}
	}
	n, err := w.Write(buf.Bytes())
	return n, err
}

// WriteSample24 -
func (w *Writer) WriteSample24(samples []byte) (int, error) {
	_ = samples
	return 0, fmt.Errorf("WriteSample24 is not implemented")
}

// Write -
func (w *Writer) Write(p []byte) (int, error) {
	blockSize := int(w.fmtChunk.Data.BlockSize)
	if len(p) < blockSize {
		return 0, fmt.Errorf("writing data need at least %d bytes", blockSize)
	}
	// 写入byte数是BlockSize的倍数
	if len(p)%blockSize != 0 {
		return 0, fmt.Errorf("writing data must be a multiple of %d bytes", blockSize)
	}
	num := len(p) / blockSize

	// 缓存超5秒数据，写盘
	if w.writtenSamples > int(w.fmtChunk.Data.SamplesPerSec)*5 {
		if err := w.flush(); err != nil {
			return 0, err
		}
	}

	n, err := w.dataChunk.Data.Write(p)
	if err == nil {
		w.writtenSamples += num
	}

	return n, err
}

type errWriter struct {
	w   io.Writer
	err error
}

// Write -
func (ew *errWriter) Write(order binary.ByteOrder, data interface{}) {
	if ew.err != nil {
		return
	}
	ew.err = binary.Write(ew.w, order, data)
}

func (w *Writer) flush() error {

	data := w.dataChunk.Data.Bytes()
	dataSize := uint32(len(data))

	if w.flushCount == 0 {
		w.riffChunk.Size = uint32(len(w.riffChunk.ID)) + (8 + w.fmtChunk.Size) + (8 + dataSize)
		w.dataChunk.Size = dataSize

		// 写wave头信息
		hdata, err := w.HeaderBytes()
		if err != nil {
			return err
		}

		if _, err := w.out.Write(hdata); err != nil {
			return err
		}
	} else {
		w.riffChunk.Size += dataSize
		w.dataChunk.Size += dataSize
	}

	// 写数据
	if _, err := w.out.Write(data); err != nil {
		return err
	}

	w.dataChunk.Data.Reset()
	bs := make([]byte, 4)

	// riff chunk size
	binary.LittleEndian.PutUint32(bs, w.riffChunk.Size)
	if _, err := w.out.WriteAt(bs, 4); err != nil {
		return err
	}

	// data chunk size
	binary.LittleEndian.PutUint32(bs, w.dataChunk.Size)
	if _, err := w.out.WriteAt(bs, 40); err != nil {
		return err
	}

	w.flushCount++
	w.writtenSamples = 0
	return nil
}

// HeaderBytes -
func (w *Writer) HeaderBytes() ([]byte, error) {
	out := bytes.NewBuffer([]byte{})

	ew := &errWriter{w: out}

	// riff chunk
	ew.Write(binary.BigEndian, w.riffChunk.ID)         // 4
	ew.Write(binary.LittleEndian, w.riffChunk.Size)    // 4
	ew.Write(binary.BigEndian, w.riffChunk.FormatType) // 4

	// fmt chunk
	ew.Write(binary.BigEndian, w.fmtChunk.ID)      // 4
	ew.Write(binary.LittleEndian, w.fmtChunk.Size) // 4
	ew.Write(binary.LittleEndian, w.fmtChunk.Data) // 16

	//data chunk
	ew.Write(binary.BigEndian, w.dataChunk.ID)      // 4
	ew.Write(binary.LittleEndian, w.dataChunk.Size) // 4

	if ew.err != nil {
		return nil, ew.err
	}

	return out.Bytes(), nil
}

// Close -
func (w *Writer) Close() error {
	if w.out == nil {
		return fmt.Errorf("not find file handle")
	}

	if err := w.flush(); err != nil {
		return err
	}

	return w.out.Close()
}

// CreateWaveHeader -
func CreateWaveHeader(samplerate int, channelCount int, dataSize uint32) ([]byte, error) {
	w, err := NewBuffer(samplerate, channelCount)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	w.riffChunk.Size = uint32(len(w.riffChunk.ID)) + (8 + w.fmtChunk.Size) + (8 + dataSize)
	w.dataChunk.Size = dataSize
	return w.HeaderBytes()
}

// PCMToWave -
//
//goland:noinspection GoUnusedExportedFunction
func PCMToWave(samplerate int, channelCount int, pcm []byte) ([]byte, error) {
	hd, err := CreateWaveHeader(samplerate, channelCount, uint32(len(pcm)))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return append(hd, pcm...), nil
}
