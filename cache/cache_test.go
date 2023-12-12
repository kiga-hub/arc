package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/kiga-hub/arc/logging"
	"github.com/kiga-hub/arc/logging/conf"
	microConf "github.com/kiga-hub/arc/micro/conf"
)

func Test_test(t *testing.T) {
	logger, err := logging.CreateLogger(&microConf.BasicConfig{
		Zone:       "Prod1",
		AppVersion: "v1.0.0",
		AppName:    "test_graylog",
	}, &conf.LogConfig{
		Level:       "DEBUG",
		Path:        "",
		GraylogAddr: "192.168.1.1:12201",
		Console:     true,
	})
	if err != nil {
		t.Errorf("%v", err)
	}
	l := logger.Sugar()

	tests := []struct {
		name string // test name
		id   uint64 // test id
		data []byte // data
		err  error
	}{
		{
			id:   1,
			name: "test_continuous_data#0",
			data: make([]byte, 2048),
			err:  nil,
		},
		{
			id:   1,
			name: "test_nonContinuous_data#1",
			data: make([]byte, 2048),
			err:  nil,
		},
		{
			id:   1,
			name: "test_nonContinuous_data#2",
			data: make([]byte, 2048),
			err:  nil,
		},
		{
			id:   1,
			name: "test_nonContinuous_data#3",
			data: make([]byte, 2048),
			err:  nil,
		},
		{
			id:   2,
			name: "test_continuous_data#4",
			data: make([]byte, 2048),
			err:  nil,
		},
		{
			id:   2,
			name: "test_nonContinuous_data#5",
			data: make([]byte, 2048),
			err:  nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cacheContainer := NewDataCacheContainer(2, 120000, true, l)
			cacheContainer.Start(true)
			startT := time.Now()
			inputT := startT
			points := float64(len(test.data))
			interval := int64(points) * 1e6

			for frameCount := 1; frameCount < len(test.data); frameCount++ {
				t.Log("frameCount", frameCount)
				cacheContainer.Input(&DataPoint{
					ID:   test.id,
					Time: inputT,
					Data: test.data,
				})

				// 设置每帧数据的时间戳
				inputT = inputT.Add(time.Duration(interval) * time.Microsecond)
			}

			// 查询所有数据，计算查询开始结束时间
			timeFrom := startT
			timeTo := timeFrom.Add(time.Duration(interval) * time.Microsecond)

			dataPoints, err := cacheContainer.Search(&SearchRequest{
				ID:       test.id,
				TimeFrom: timeFrom.UnixMicro(),
				TimeTo:   timeTo.UnixMicro(),
			})
			if err != test.err {
				t.Fatalf("cacheSearch %v %T", err, timeFrom)
			}

			if dataPoints[0] == nil {
				stat := cacheContainer.GetStat()
				if stat != nil {
					t.Log("The query time is within the cache time range |dataLen<1|stat!=nil",
						"sensorID-int", test.id,
						"currentSystemTime", time.Now().UTC(),
						"stat size", stat[test.id].Size,
						"stat from", stat[test.id].From,
						"stat to", stat[test.id].To,
						"stat expire", stat[test.id].Expire,
						"stat count", stat[test.id].Count,
					)
				}
				t.Fatalf("dataPoints is null %T", dataPoints[0])
			}
			data := dataPoints[0].(*DataPoint)

			if len(data.Data) != len(test.data) {
				outputFunc(test.id, dataPoints)
				t.Fatalf("dataSize %d-%d", len(data.Data), len(test.data))
			}

			t.Logf("len(dataPoints): %d", len(dataPoints))

			for _, v := range dataPoints {
				point := v.(*DataPoint)
				// 比较查询数据与预期数据大小，缓存读取数据大小，小于等于每帧数据大小*(包序号取模-1)
				t.Log("Expected", len(point.Data), len(test.data))
				if len(point.Data) > len(test.data) {
					outputFunc(test.id, dataPoints)
					t.Fatalf("dataSize %d-%d", len(point.Data), len(test.data))
				}
			}

			time.Sleep(time.Second * 1)
			cacheContainer.Stop()
		})
	}

}

func outputFunc(id uint64, data []IDataPoint) {
	for _, point := range data {
		if point == nil {
			continue
		}
		fmt.Printf("id, %d\n", id)
		fmt.Printf("getID, %d:%d\n", id, point.GetID())
		fmt.Printf("getTime, %d:%d\n", id, point.GetTime())
		fmt.Printf("getSize, %d:%d\n", id, point.GetSize())
	}
}
