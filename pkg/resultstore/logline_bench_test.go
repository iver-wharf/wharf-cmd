package resultstore

import (
	"io"
	"io/fs"
	"sync"
	"testing"
)

var (
	sampleShortLogLine = sampleTimeStr + " Build done."
	sampleLongLogLine  = sampleTimeStr + " Some long log line that comes from the Wharf build logs that spans a really long line that could be a something like multiple exception messages all grouped together in a single line"
)

func BenchmarkLogChanShort0(b *testing.B) {
	const buffer = 0
	benchmarkLogChan(b, buffer, sampleShortLogLine)
}

func BenchmarkLogChanShort1(b *testing.B) {
	const buffer = 1
	benchmarkLogChan(b, buffer, sampleShortLogLine)
}

func BenchmarkLogChanShort100(b *testing.B) {
	const buffer = 100
	benchmarkLogChan(b, buffer, sampleShortLogLine)
}

func BenchmarkLogChanShort1000(b *testing.B) {
	const buffer = 10000
	benchmarkLogChan(b, buffer, sampleShortLogLine)
}

func BenchmarkLogChanLong0(b *testing.B) {
	const buffer = 0
	benchmarkLogChan(b, buffer, sampleLongLogLine)
}

func BenchmarkLogChanLong1(b *testing.B) {
	const buffer = 1
	benchmarkLogChan(b, buffer, sampleLongLogLine)
}

func BenchmarkLogChanLong100(b *testing.B) {
	const buffer = 100
	benchmarkLogChan(b, buffer, sampleLongLogLine)
}

func BenchmarkLogChanLong1000(b *testing.B) {
	const buffer = 10000
	benchmarkLogChan(b, buffer, sampleLongLogLine)
}

func benchmarkLogChan(b *testing.B, buffer int, line string) {
	const stepID uint64 = 1
	s := NewStore(mockFS{
		listDirEntries: func(name string) ([]fs.DirEntry, error) {
			return nil, nil
		},
		openRead: func(name string) (io.ReadCloser, error) {
			return nil, fs.ErrNotExist
		},
		openAppend: func(name string) (io.WriteCloser, error) {
			return nopWriteCloser{}, nil
		},
	})
	w, err := s.OpenLogWriter(stepID)
	if err != nil {
		b.Fatal(err)
	}
	ch, err := s.SubAllLogLines(buffer)
	if err != nil {
		b.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for logLine := range ch {
			// read all, but do nothing with the values
			_ = logLine
		}
		wg.Done()
	}()
	b.SetBytes(int64(len(line)))
	for n := 0; n < b.N; n++ {
		w.WriteLogLine(line)
	}
	s.UnsubAllLogLines(ch)
	wg.Wait()
}
