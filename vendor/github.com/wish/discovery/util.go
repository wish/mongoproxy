package discovery

import "math/rand"

// ShuffleServiceAddresses will randomize ServiceAddresses
func ShuffleServiceAddresses(arr ServiceAddresses) ServiceAddresses {
	rand.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})

	return arr
}
