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
	WaveFormatType uint16 // PCM 为 1
	Channel        uint16 // 声道
	SamplesPerSec  uint32 // 采样频率 44100
	BytesPerSec    uint32 // 1秒所需要的byte数
	BlockSize      uint16 // 量子化精度 * 通道数
	BitsPerSamples uint16 // 量子化精度
}

// DataReader - 装入
type DataReader interface {
	io.Reader
	io.ReaderAt
}

// DataReaderChunk -
type DataReaderChunk struct {
	ID   []byte     // 'data'
	Size uint32     // 声音数据长度 * channel
	Data DataReader // 实际数据
}

// DataWriterChunk -
type DataWriterChunk struct {
	ID   []byte
	Size uint32
	Data *bytes.Buffer
}
