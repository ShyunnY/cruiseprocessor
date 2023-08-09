package cruiseprocessor

import (
	"crypto/sha1"
	"hash/fnv"
	"math"
	"math/big"
	"os"
)

func hasSampling(traceID []byte, salt string, percentage uint64) bool {
	hash := fnv.New64a()

	// we don't have to handle that error.
	// Write() will never return an error
	hash.Write(traceID)
	hash.Write([]byte(salt))

	return hash.Sum64() < percentage
}

func computeSampleRate(percentage float64) uint64 {
	b := new(big.Float).SetInt(new(big.Int).SetUint64(math.MaxUint64))
	p, _ := b.Mul(b, big.NewFloat(percentage/100)).Uint64()
	return p
}

func defaultSaltVal() string {
	hostname, err := os.Hostname()
	if err != nil {
		// TODO: consider whether to use timestamps
	}

	hash := sha1.New()
	hash.Write([]byte(hostname))
	identity := hash.Sum(nil)

	return string(identity)
}
