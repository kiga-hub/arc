package wave

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/pkg/errors"
)

// CutReader -
type CutReader struct {
	fr       *os.File // 文件
	finfo    fs.FileInfo
	foffset  int64         // 文件偏移位置
	fpoint   int64         // 采样点数量
	fchannel int           // 文件声道数
	fbits    int           // 文件位深度
	hr       io.ReadSeeker // 切分文件头信息
	hcurrent int           // 当前读取头信息位置
	channels []int         // 切分文件通道下标
	poffset  int           // 切分文件偏移位置
	cpoint   int64         // 当前读取采样点数量
}

// NewCutReaderByMsec -
func NewCutReaderByMsec(fileName string, offsetMs, durationMs int, channels []int) (*CutReader, error) {
	// 打开文件
	f, err := os.Open(fileName)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// 获取文件属性信息
	d, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, errors.WithStack(err)
	}

	// 读取wave获取头
	header := make([]byte, 44)
	n, err := f.Read(header)
	if err != nil {
		f.Close()
		return nil, errors.WithStack(err)
	}
	if n != 44 {
		f.Close()
		return nil, errors.New("read wave header not enough 44 bytes")
	}

	// 检查通道数
	channel := binary.LittleEndian.Uint16(header[22:])
	for _, cidx := range channels {
		if cidx >= int(channel) || cidx < 0 {
			return nil, errors.Errorf("%s invalid channels %d", fileName, cidx)
		}
	}

	if len(channels) == 0 {
		for i := 0; i < int(channel); i++ {
			channels = append(channels, i)
		}
	}

	// 获取采样率
	samplerate := binary.LittleEndian.Uint32(header[24:])
	// 平均字节速率
	nAvgBytesPerSec := binary.LittleEndian.Uint32(header[28:])
	// 计算文件音频时长(毫秒)
	totalDurationMs := uint32((d.Size() - 44) * 1000 / int64(nAvgBytesPerSec))
	// 位深度(字节)
	bits := nAvgBytesPerSec / samplerate / uint32(channel)

	// 检查时间编译
	if int(totalDurationMs) <= offsetMs {
		return nil, errors.Errorf("%s invalid offset time %d <= %d ", fileName, totalDurationMs, offsetMs)
	}
	if int(totalDurationMs)-offsetMs < durationMs {
		durationMs = int(totalDurationMs) - offsetMs
	}
	if durationMs <= 0 {
		durationMs = int(totalDurationMs) - offsetMs
	}

	// 文件偏移位置
	offset := (offsetMs*int(samplerate))/1000*int(bits)*int(channel) + 44

	// 目标文件大小(不包括头44字节)
	size := (uint32(durationMs) * samplerate) / 1000 * bits * uint32(len(channels))

	// 定位
	if _, err := f.Seek(int64(offset), 0); err != nil {
		return nil, errors.WithStack(err)
	}

	// 创建头信息
	hd, err := CreateWaveHeader(int(samplerate), len(channels), size)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &CutReader{
		hr:       bytes.NewReader(hd),
		fr:       f,
		finfo:    d,
		foffset:  int64(offset),
		fpoint:   (int64(durationMs) * int64(samplerate)) / 1000,
		fchannel: int(channel),
		fbits:    int(bits),
		channels: channels,
	}, nil
}

// NewCutReaderByTime -
func NewCutReaderByTime(fileName string, offset, duration time.Duration, channels []int) (*CutReader, error) {
	// 打开文件
	f, err := os.Open(fileName)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// 获取文件属性信息
	d, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, errors.WithStack(err)
	}

	// 读取wave获取头
	header := make([]byte, 44)
	n, err := f.Read(header)
	if err != nil {
		f.Close()
		return nil, errors.WithStack(err)
	}
	if n != 44 {
		f.Close()
		return nil, errors.New("read wave header not enough 44 bytes")
	}

	// 检查通道数
	channel := binary.LittleEndian.Uint16(header[22:])
	for _, cidx := range channels {
		if cidx >= int(channel) || cidx < 0 {
			return nil, errors.Errorf("%s invalid channels %d", fileName, cidx)
		}
	}

	if len(channels) == 0 {
		for i := 0; i < int(channel); i++ {
			channels = append(channels, i)
		}
	}

	// 获取采样率
	samplerate := binary.LittleEndian.Uint32(header[24:])
	// 平均字节速率
	nAvgBytesPerSec := binary.LittleEndian.Uint32(header[28:])
	// 计算文件音频时长
	totalDuration := time.Duration(float64(d.Size()-44) / float64(nAvgBytesPerSec) * float64(time.Second))
	// 位深度(字节)
	bits := nAvgBytesPerSec / samplerate / uint32(channel)

	// 检查时间编译
	if totalDuration <= offset {
		return nil, errors.Errorf("%s invalid offset time %s <= %s ", fileName, totalDuration.String(), offset.String())
	}
	if duration <= 0 || totalDuration-offset < duration {
		duration = totalDuration - offset
	}

	// 文件偏移位置
	foffset := (offset.Milliseconds()*int64(samplerate))/1000*int64(bits)*int64(channel) + 44

	// 目标文件大小(不包括头44字节)
	size := uint32((duration.Milliseconds()*int64(samplerate))/1000) * bits * uint32(len(channels))

	// 定位
	if _, err := f.Seek(int64(foffset), 0); err != nil {
		return nil, errors.WithStack(err)
	}

	// 创建头信息
	hd, err := CreateWaveHeader(int(samplerate), len(channels), size)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &CutReader{
		hr:       bytes.NewReader(hd),
		fr:       f,
		finfo:    d,
		foffset:  foffset,
		fpoint:   duration.Milliseconds() * int64(samplerate) / 1000,
		fchannel: int(channel),
		fbits:    int(bits),
		channels: channels,
	}, nil
}

// NewCutReader -
func NewCutReader(fileName string, offset, size int64) (*CutReader, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	d, err := f.Stat()
	if err != nil {
		return nil, err
	}

	// 获取头
	header := make([]byte, 44)
	n, err := f.Read(header)
	if err != nil {
		return nil, err
	}
	if n != 44 {
		return nil, fmt.Errorf("read wave header not enough 44 bytes")
	}
	channel := binary.LittleEndian.Uint16(header[22:])
	// 获取采样率
	samplerate := binary.LittleEndian.Uint32(header[24:])
	nAvgBytesPerSec := binary.LittleEndian.Uint32(header[28:])
	// 位深度(字节)
	bits := nAvgBytesPerSec / samplerate / uint32(channel)

	var channels []int
	for i := 0; i < int(channel); i++ {
		channels = append(channels, i)
	}

	// 定位
	if _, err := f.Seek(offset, 0); err != nil {
		return nil, err
	}

	// 获取头
	hd, err := CreateWaveHeader(int(samplerate), int(channel), uint32(size))
	if err != nil {
		return nil, err
	}
	wr := &CutReader{
		hr:       bytes.NewReader(hd),
		fr:       f,
		finfo:    d,
		foffset:  offset,
		fpoint:   size / (int64(channel) * int64(bits)),
		fchannel: int(channel),
		fbits:    int(bits),
		channels: channels,
	}
	return wr, nil
}

// Read -
func (rd *CutReader) Read(p []byte) (n int, err error) {
	// 判断是否读到切分位置
	if rd.fpoint == 0 || rd.fpoint <= rd.cpoint {
		return 0, io.EOF
	}

	// 读头信息
	n, err = rd.hr.Read(p)
	rd.hcurrent += n
	left := len(p) - n
	if left <= 0 {
		return
	}

	var point int
	var poffset int

	// 切分文件采样点大小
	psize := len(rd.channels) * rd.fbits
	// 切分文件剩余大小
	sizeLeft := (rd.fpoint - rd.cpoint) * int64(len(rd.channels)) * int64(rd.fbits)

	if sizeLeft < int64(left) {
		/*
			剩余数据一次性读完, 计算读取采样点数
		*/
		left = int(sizeLeft)
		point = left / psize
		rd.cpoint = rd.fpoint
	} else {
		/*
			剩余数据一次性不能读完, 计算读取采样点数
		*/
		if rd.poffset > 0 {
			// 上次读取存在不完整采样点
			left -= psize - rd.poffset
			point++
			rd.cpoint++
		}

		// 剩余空间存在不完成采样点
		poffset = left % psize
		if poffset > 0 {
			left -= poffset
			point++
		}

		// 完整采样点
		point += left / psize
		rd.cpoint += int64(point)
	}

	// 按照文件通道数从读取采样点
	buf := make([]byte, point*rd.fchannel*rd.fbits)
	_, err = rd.fr.Read(buf)
	if err != nil {
		return 0, err
	}

	// 如果不是整采样点，文件读取位置后退一个采样点
	if poffset > 0 {
		_, err = rd.fr.Seek(int64(-1*rd.fchannel*rd.fbits), io.SeekCurrent)
		if err != nil {
			return 0, err
		}
	}

	defer func() {
		rd.poffset = poffset
	}()

	// 拼接目标数据
	for i := 0; i < len(buf); i += rd.fchannel * rd.fbits {
		for _, cidx := range rd.channels {
			for j := 0; j < rd.fbits; j++ {
				if rd.poffset > 0 {
					rd.poffset--
					continue
				}
				p[n] = buf[i+(cidx*rd.fbits)+j]
				n++
				if n >= len(p) {
					return
				}
			}
		}
	}

	return
}

func (rd *CutReader) fSeek(offset int64) (int64, error) {
	rd.cpoint = (offset - rd.foffset) / (int64(rd.fchannel * rd.fbits))
	return rd.fr.Seek(offset, io.SeekStart)
}

func (rd *CutReader) hSeek(offset int64) (int64, error) {
	rd.hcurrent = int(offset)
	return rd.hr.Seek(offset, io.SeekStart)
}

// Seek - 只为 http.ServeContent 调用使用
func (rd *CutReader) Seek(offset int64, whence int) (int64, error) {
	psize := int64(len(rd.channels) * rd.fbits)
	total := rd.fpoint * psize

	var abs int64
	switch whence {
	case io.SeekCurrent:
		offset += rd.cpoint*psize + int64(rd.poffset) + int64(rd.hcurrent)
	case io.SeekEnd:
		offset += total + 44
	}
	if offset <= 44 {
		if _, err := rd.fSeek(rd.foffset); err != nil {
			return 0, err
		}
		return rd.hSeek(offset)
	}
	rd.poffset = (int(offset) - 44) % (len(rd.channels) * rd.fbits)
	offset -= 44 + int64(rd.poffset)
	if len(rd.channels) != rd.fchannel {
		offset = offset / int64(len(rd.channels)) * int64(rd.fchannel)
	}
	abs = offset + rd.foffset

	if _, err := rd.hSeek(44); err != nil {
		return 0, err
	}

	if abs > total+rd.foffset {
		abs = total + rd.foffset
	}

	n, err := rd.fSeek(abs)
	if err != nil {
		return 0, err
	}
	return n - rd.foffset + 44, nil
}

// Close -
func (rd *CutReader) Close() error {
	return rd.fr.Close()
}

// Name -
func (rd *CutReader) Name() string {
	return rd.finfo.Name()
}

// ModTime -
func (rd *CutReader) ModTime() time.Time {
	return rd.finfo.ModTime()
}
