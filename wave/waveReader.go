package wave

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// waveReader -
type waveReader interface {
	io.Reader
	io.Seeker
	io.ReaderAt
}

// Reader -
type Reader struct {
	input waveReader

	size int64

	RiffChunk *RiffChunk
	FmtChunk  *FmtChunk
	DataChunk *DataReaderChunk

	originOfAudioData int64
	NumSamples        uint32
	ReadSampleNum     uint32
	SampleTime        float64

	// Used to manage variables of variable chunk length such as LIST chunk"
	extChunkSize int64
}

// NewReader -
//
//goland:noinspection GoUnusedExportedFunction
func NewReader(fileName string) (*Reader, error) {
	// check file size
	fi, err := os.Stat(fileName)
	if err != nil {
		return &Reader{}, err
	}
	if fi.Size() > maxFileSize {
		return &Reader{}, fmt.Errorf("file is too large: %d bytes", fi.Size())
	}

	// open file
	f, err := os.Open(fileName)
	if err != nil {
		return &Reader{}, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	waveData, err := io.ReadAll(f)
	if err != nil {
		return &Reader{}, err
	}

	return NewReaderByData(waveData, fi.Size())
}

// NewReaderByFile -
//
//goland:noinspection GoUnusedExportedFunction
func NewReaderByFile(fileName string) (*Reader, error) {
	// check file size
	fi, err := os.Stat(fileName)
	if err != nil {
		return nil, err
	}
	if fi.Size() > maxFileSize {
		return nil, fmt.Errorf("file is too large: %d bytes", fi.Size())
	}

	// open file
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	reader := new(Reader)
	reader.size = fi.Size()
	reader.input = f

	if err := reader.parseRiffChunk(); err != nil {
		return nil, err
	}
	if err := reader.parseFmtChunk(); err != nil {
		return nil, err
	}
	if err := reader.parseListChunk(); err != nil {
		return nil, err
	}
	if err := reader.parseDataChunk(); err != nil {
		return nil, err
	}

	reader.NumSamples = reader.DataChunk.Size / uint32(reader.FmtChunk.Data.BlockSize)
	reader.SampleTime = float64(reader.NumSamples) / float64(reader.FmtChunk.Data.SamplesPerSec)

	return reader, nil
}

// NewReaderByData -
func NewReaderByData(waveData []byte, size int64) (*Reader, error) {
	reader := new(Reader)
	reader.size = size
	reader.input = bytes.NewReader(waveData)

	if err := reader.parseRiffChunk(); err != nil {
		panic(err)
	}
	if err := reader.parseFmtChunk(); err != nil {
		panic(err)
	}
	if err := reader.parseListChunk(); err != nil {
		panic(err)
	}
	if err := reader.parseDataChunk(); err != nil {
		panic(err)
	}

	reader.NumSamples = reader.DataChunk.Size / uint32(reader.FmtChunk.Data.BlockSize)
	reader.SampleTime = float64(reader.NumSamples) / float64(reader.FmtChunk.Data.SamplesPerSec)

	return reader, nil
}

type cSize struct {
	ChunkSize uint32
}

func (rd *Reader) parseRiffChunk() error {
	// RIFF format header check
	chunkID := make([]byte, 4)
	if err := binary.Read(rd.input, binary.BigEndian, chunkID); err != nil {
		return err
	}
	if string(chunkID[:]) != riffChunkToken {
		return fmt.Errorf("file is not RIFF: %s", rd.RiffChunk.ID)
	}

	// RIFF information block size
	chunkSize := &cSize{}
	if err := binary.Read(rd.input, binary.LittleEndian, chunkSize); err != nil {
		return err
	}
	if chunkSize.ChunkSize+8 != uint32(rd.size) {
		//		fmt.Println("======================")
		//		fmt.Println("riff chunk size ", rd.riffChunk.ChunkSize)
		//		fmt.Println("file size ", rd.size)
		//		fmt.Println("======================")
		return fmt.Errorf("riff_chunk_size must be whole file size - 8bytes, expected(%d), actual(%d)", chunkSize.ChunkSize+8, rd.size)
	}

	// Whether there is a RIFF format data type check 'WAVE
	format := make([]byte, 4)
	if err := binary.Read(rd.input, binary.BigEndian, format); err != nil {
		return err
	}
	if string(format[:]) != waveFormatType {
		return fmt.Errorf("file is not WAVE: %s", rd.RiffChunk.FormatType)
	}

	riffChunk := RiffChunk{
		ID:         chunkID,
		Size:       chunkSize.ChunkSize,
		FormatType: format,
	}

	rd.RiffChunk = &riffChunk

	return nil
}

func (rd *Reader) parseFmtChunk() error {
	if _, err := rd.input.Seek(riffChunkSize, io.SeekStart); err != nil {
		return err
	}

	// ‘fmt’
	chunkID := make([]byte, 4)
	err := binary.Read(rd.input, binary.BigEndian, chunkID)
	if err == io.EOF {
		return fmt.Errorf("unexpected file end")
	} else if err != nil {
		return err
	}
	if string(chunkID[:]) != fmtChunkToken {
		return fmt.Errorf("fmt chunk id must be \"%s\" but value is %s", fmtChunkToken, chunkID)
	}

	// fmt chunk size i 16 bit
	chunkSize := &cSize{}
	err = binary.Read(rd.input, binary.LittleEndian, chunkSize)
	if err == io.EOF {
		return fmt.Errorf("unexpected file end")
	} else if err != nil {
		return err
	}
	if chunkSize.ChunkSize != fmtChunkSize {
		return fmt.Errorf("fmt chunk size must be %d but value is %d", fmtChunkSize, chunkSize.ChunkSize)
	}

	// read fmt chunk data
	var fmtChunkData WavFmtChunkData
	if err = binary.Read(rd.input, binary.LittleEndian, &fmtChunkData); err != nil {
		return err
	}

	fmtChunk := FmtChunk{
		ID:   chunkID,
		Size: chunkSize.ChunkSize,
		Data: &fmtChunkData,
	}

	rd.FmtChunk = &fmtChunk

	return nil
}

func (rd *Reader) parseListChunk() error {
	if _, err := rd.input.Seek(listChunkOffset, io.SeekStart); err != nil {
		return err
	}

	// “LIST”
	chunkID := make([]byte, 4)
	if err := binary.Read(rd.input, binary.BigEndian, chunkID); err == io.EOF {
		return fmt.Errorf("unexpected file end")
	} else if err != nil {
		return err
	} else if string(chunkID[:]) != listChunkToken {
		//
		return nil
	}

	chunkSize := make([]byte, 1)
	if err := binary.Read(rd.input, binary.LittleEndian, chunkSize); err == io.EOF {
		return fmt.Errorf("unexpected file end")
	} else if err != nil {
		return err
	}

	// Update of variable header length management variable
	// rd.extChunkSize += int64(chunkSize[0]) + 4 + 4
	rd.extChunkSize = int64(chunkSize[0]) + 4 + 4

	return nil
}

func (rd *Reader) getRiffChunkSizeOffset() int64 {
	return riffChunkSizeBaseOffset + rd.extChunkSize
}

func (rd *Reader) parseDataChunk() error {
	originOfDataChunk, _ := rd.input.Seek(rd.getRiffChunkSizeOffset(), io.SeekStart)

	chunkID := make([]byte, 4)
	err := binary.Read(rd.input, binary.BigEndian, chunkID)
	if err == io.EOF {
		return fmt.Errorf("unexpected file end")
	} else if err != nil {
		return err
	}
	if string(chunkID[:]) != dataChunkToken {
		return fmt.Errorf("data chunk id must be \"%s\" but value is %s", dataChunkToken, chunkID)
	}

	// data_chunk_size
	chunkSize := &cSize{}
	err = binary.Read(rd.input, binary.LittleEndian, chunkSize)
	if err == io.EOF {
		return fmt.Errorf("unexpected file end")
	} else if err != nil {
		return err
	}

	rd.originOfAudioData = originOfDataChunk + 8
	audioData := io.NewSectionReader(rd.input, rd.originOfAudioData, int64(chunkSize.ChunkSize))

	dataChunk := DataReaderChunk{
		ID:   chunkID,
		Size: chunkSize.ChunkSize,
		Data: audioData,
	}

	rd.DataChunk = &dataChunk

	return nil
}

// onlu read sound data
func (rd *Reader) Read(p []byte) (int, error) {
	n, err := rd.DataChunk.Data.Read(p)
	return n, err
}

// ReadRawSample -
func (rd *Reader) ReadRawSample() ([]byte, error) {
	size := rd.FmtChunk.Data.BlockSize
	sample := make([]byte, size)
	_, err := rd.Read(sample)
	if err == nil {
		rd.ReadSampleNum++
	}
	return sample, err
}

// ReadSample -
func (rd *Reader) ReadSample() ([]float64, error) {
	raw, err := rd.ReadRawSample()
	channel := int(rd.FmtChunk.Data.Channel)
	ret := make([]float64, channel)
	length := len(raw) / channel // The number of bytes per channel

	if err != nil {
		return ret, err
	}

	for i := 0; i < channel; i++ {
		tmp := bytesToInt(raw[length*i : length*(i+1)])
		switch rd.FmtChunk.Data.BitsPerSamples {
		case 8:
			ret[i] = float64(tmp-128) / 128.0
		case 16:
			ret[i] = float64(tmp) / 32768.0
		}
		if err != nil && err != io.EOF {
			return ret, err
		}
	}
	return ret, nil
}

// ReadSampleInt -
func (rd *Reader) ReadSampleInt() ([]int, error) {
	raw, err := rd.ReadRawSample()
	channels := int(rd.FmtChunk.Data.Channel)
	ret := make([]int, channels)
	length := len(raw) / channels // The number of bytes per channel

	if err != nil {
		return ret, err
	}

	for i := 0; i < channels; i++ {
		ret[i] = bytesToInt(raw[length*i : length*(i+1)])
		if err != nil && err != io.EOF {
			return ret, err
		}
	}
	return ret, nil
}

func bytesToInt(b []byte) int {
	var ret int
	switch len(b) {
	case 1:
		// 0 ~ 128 ~ 255
		ret = int(b[0])
	case 2:
		// -32768 ~ 0 ~ 32767
		ret = int(b[0]) + int(b[1])<<8
	//	fmt.Printf("%08b %08b ", b[1], b[0])
	//	fmt.Printf("%016b => %d\n", ret, ret)
	case 3:
		// Hi-Re-so / DVDAudio
		ret = int(b[0]) + int(b[1])<<8 + int(b[2])<<16
	default:
		ret = 0
	}
	return ret
}
