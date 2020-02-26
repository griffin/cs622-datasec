package util

import (
	"math/rand"
	"sync"
	"time"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var (
	randLk  = &sync.Mutex{}
	randSrc = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func randInt63() (n int64) {
	randLk.Lock()
	n = randSrc.Int63()
	randLk.Unlock()
	return
}

// GenerateSelector generates a selector of length n using the env's random
// Safe for concurrent use.
func GenerateSelector(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, randInt63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = randInt63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
