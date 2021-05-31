package utils

import "fmt"

func GetImage(imageName string, separator rune, version string) string {
	return fmt.Sprintf("%s%c%s", imageName, separator, version)
}