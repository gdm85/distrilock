package dlclient

import (
	"fmt"
	"math/rand"
	"testing"
)

func BenchmarkLocksTaking(b *testing.B) {
	fixedLockName := fmt.Sprintf("testing-%d", rand.Int())
	type nameGen func() string
	benchmarks := []nameGen{
		func() string {
			return fmt.Sprintf("testing-%d", rand.Int())
		},
		func() string {
			return fixedLockName
		},
	}
	for i, bm := range benchmarks {
		b.Run(fmt.Sprintf("benchmark#%d", i), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				lockName := bm()
				l, err := testClientA1.Acquire(lockName)
				if err != nil {
					b.Error(err)
					return
				}
				err = l.Release()
				if err != nil {
					b.Error(err)
					return
				}
			}
		})
	}
}
