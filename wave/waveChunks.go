package wave

import (
	"bytes"
	"io"
)

const (
	maxFileSize             = 2 << 31
	riffChunkSize           = 12
	listChunkOffset         = 36
	riffChunkSizeBaseOffset = 36 // RIFFChunk(12byte) + fmtChunk(24byte) = 36byte
	fmtChunkSize            = 16
)

var (
	riffChunkToken = "RIFF"
	waveFormatType = "WAVE"
	fmtChunkToken  = "fmt "
	listChunkToken = "LIST"
	dataChunkToken = "data"
)

// RiffChunk - 12byte
type RiffChunk struct {
	ID         []byte // 'RIFF'
	Size       uint32 // 36bytes + data_chunk_size or whole_file_size - 'RIFF'+ChunkSize (8byte)
	FormatType []byte // 'WAVE'
}

// FmtChunk - 8 + 16 = 24byte
type FmtChunk struct {
	ID   []byte // 'fmt '
	Size uint32 // 16
	Data *WavFmtChunkData
}

// WavFmtChunkData - 6byte
type WavFmtChunkData struct {
	WaveFormatType uint16 // PCM  1
	Channel        uint16 // channel
	SamplesPerSec  uint32
	BytesPerSec    uint32 // The number of bytes required per second
	BlockSize      uint16 // Quantization accuracy * number of channels
	BitsPerSamples uint16 // Quantization accuracy
}

// DataReader - load
type DataReader interface {
	io.Reader
	io.ReaderAt
}

// DataReaderChunk -
type DataReaderChunk struct {
	ID   []byte     // 'data'
	Size uint32     // data size * channel
	Data DataReader // actual data
}

// DataWriterChunk -
type DataWriterChunk struct {
	ID   []byte
	Size uint32
	Data *bytes.Buffer
}
