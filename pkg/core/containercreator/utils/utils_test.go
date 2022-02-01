package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetImage(t *testing.T) {
	imageName := "alpine/git"
	separator := ':'
	version := "v2.30.1"

	imageNameWithVersion := GetImage(imageName, separator, version)
	require.Equal(t, "alpine/git:v2.30.1", imageNameWithVersion)
}
