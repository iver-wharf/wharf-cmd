package slices

// ReverseStrings will reverse the order of all elements in the slice in place,
// meaning it will alter the existing slice.
func ReverseStrings(slice []string) {
	Reverse(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}

// Reverse will reverse the order of all elements in the slice using a swap
// function that should swap two elements in the slice.
func Reverse(length int, swap func(i, j int)) {
	for i, j := 0, length-1; i < length/2; i, j = i+1, j-1 {
		swap(i, j)
	}
}
