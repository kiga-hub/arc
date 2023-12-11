package wave

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
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

	// 用于管理诸如LIST chunk之类的可变chunk长度的变量
	extChunkSize int64
}

// NewReader -
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
	defer f.Close()

	waveData, err := ioutil.ReadAll(f)
	if err != nil {
		return &Reader{}, err
	}

	return NewReaderByData(waveData, fi.Size())
}

// NewReaderByFile -
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
	defer f.Close()

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

type csize struct {
	ChunkSize uint32
}

func (rd *Reader) parseRiffChunk() error {
	// RIFF格式头检查
	chunkid := make([]byte, 4)
	if err := binary.Read(rd.input, binary.BigEndian, chunkid); err != nil {
		return err
	}
	if string(chunkid[:]) != riffChunkToken {
		return fmt.Errorf("file is not RIFF: %s", rd.RiffChunk.ID)
	}

	// RIFF信息块大小
	chunkSize := &csize{}
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

	// 是否写有RIFF格式数据类型检查“WAVE”
	format := make([]byte, 4)
	if err := binary.Read(rd.input, binary.BigEndian, format); err != nil {
		return err
	}
	if string(format[:]) != waveFormatType {
		return fmt.Errorf("file is not WAVE: %s", rd.RiffChunk.FormatType)
	}

	riffChunk := RiffChunk{
		ID:         chunkid,
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

	// 是否写着‘fmt’
	chunkid := make([]byte, 4)
	err := binary.Read(rd.input, binary.BigEndian, chunkid)
	if err == io.EOF {
		return fmt.Errorf("unexpected file end")
	} else if err != nil {
		return err
	}
	if string(chunkid[:]) != fmtChunkToken {
		return fmt.Errorf("fmt chunk id must be \"%s\" but value is %s", fmtChunkToken, chunkid)
	}

	// fmt chunk size是16比特吗
	chunkSize := &csize{}
	err = binary.Read(rd.input, binary.LittleEndian, chunkSize)
	if err == io.EOF {
		return fmt.Errorf("unexpected file end")
	} else if err != nil {
		return err
	}
	if chunkSize.ChunkSize != fmtChunkSize {
		return fmt.Errorf("fmt chunk size must be %d but value is %d", fmtChunkSize, chunkSize.ChunkSize)
	}

	// 读取fmt chunk data
	var fmtChunkData WavFmtChunkData
	if err = binary.Read(rd.input, binary.LittleEndian, &fmtChunkData); err != nil {
		return err
	}

	fmtChunk := FmtChunk{
		ID:   chunkid,
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

	// 是否写着“LIST”
	chunkID := make([]byte, 4)
	if err := binary.Read(rd.input, binary.BigEndian, chunkID); err == io.EOF {
		return fmt.Errorf("unexpected file end")
	} else if err != nil {
		return err
	} else if string(chunkID[:]) != listChunkToken {
		// 没有LIST信息块也没什么问题
		return nil
	}

	// ‘LIST’的尺寸是可变的，在开头的1byte中记载了尺寸
	chunkSize := make([]byte, 1)
	if err := binary.Read(rd.input, binary.LittleEndian, chunkSize); err == io.EOF {
		return fmt.Errorf("unexpected file end")
	} else if err != nil {
		return err
	}

	// 可变header长度管理变量更新
	// rd.extChunkSize += int64(chunkSize[0]) + 4 + 4
	rd.extChunkSize = int64(chunkSize[0]) + 4 + 4

	return nil
}

// 还加上可变header长度管理变量更新可变长度header尺寸的riffChunkSize Offset值
func (rd *Reader) getRiffChunkSizeOffset() int64 {
	return riffChunkSizeBaseOffset + rd.extChunkSize
}

func (rd *Reader) parseDataChunk() error {
	originOfDataChunk, _ := rd.input.Seek(rd.getRiffChunkSizeOffset(), io.SeekStart)

	// 'data' 是否写着
	chunkid := make([]byte, 4)
	err := binary.Read(rd.input, binary.BigEndian, chunkid)
	if err == io.EOF {
		return fmt.Errorf("unexpected file end")
	} else if err != nil {
		return err
	}
	if string(chunkid[:]) != dataChunkToken {
		return fmt.Errorf("data chunk id must be \"%s\" but value is %s", dataChunkToken, chunkid)
	}

	// data_chunk_size取得（实际声音数据的容量）
	chunkSize := &csize{}
	err = binary.Read(rd.input, binary.LittleEndian, chunkSize)
	if err == io.EOF {
		return fmt.Errorf("unexpected file end")
	} else if err != nil {
		return err
	}

	// 实际的声音数据是从数据Chunk的开始位置加上ID数据（4byte）和chunkSize（4byte）数据的地方
	rd.originOfAudioData = originOfDataChunk + 8
	audioData := io.NewSectionReader(rd.input, rd.originOfAudioData, int64(chunkSize.ChunkSize))

	dataChunk := DataReaderChunk{
		ID:   chunkid,
		Size: chunkSize.ChunkSize,
		Data: audioData,
	}

	rd.DataChunk = &dataChunk

	return nil
}

// 只读取声音数据
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
	length := len(raw) / channel // 每个通道的byte数

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
	length := len(raw) / channels // 每个通道的byte数

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
		// HiReso / DVDAudio
		ret = int(b[0]) + int(b[1])<<8 + int(b[2])<<16
	default:
		ret = 0
	}
	return ret
}
