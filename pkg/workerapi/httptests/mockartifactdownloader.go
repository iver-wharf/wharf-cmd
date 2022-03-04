package httptests

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

type mockArtifactDownloader struct{}

var artifactData1 = []byte("Hello, Blob 1!\r\n")
var artifactData2 = []byte("Hello, Blob 2!\r\n")
var errArtifactNotFound = errors.New("artifact not found")

func (a *mockArtifactDownloader) DownloadArtifact(artifactID uint) (io.ReadCloser, error) {
	artifacts := make(map[uint][]byte)
	artifacts[validArtifactID1] = artifactData1
	artifacts[validArtifactID2] = artifactData2
	if data, ok := artifacts[artifactID]; ok {
		return io.NopCloser(bufio.NewReader(bytes.NewBuffer(data))), nil
	}
	return nil, errArtifactNotFound
}
