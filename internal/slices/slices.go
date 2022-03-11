package slices

// ReverseStrings will reverse the order of all elements in the slice in place,
// meaning it will alter the existing slice.
func ReverseStrings(slice []string) {
	for i, j := 0, len(slice)-1; i < len(slice)/2; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}
