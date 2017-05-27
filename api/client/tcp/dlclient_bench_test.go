package dlclient

import (
	"fmt"
	"testing"
)

func BenchmarkLocksTaking(b *testing.B) {
	fixedLockName := generateLockName(b)
	type nameGen func() string
	benchmarks := []nameGen{
		func() string {
			return generateLockName(b)
		},
		func() string {
			return fixedLockName
		},
	}
	for i, bm := range benchmarks {
		b.Run(fmt.Sprintf("benchmark#%d", i), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				lockName := bm()

				c := createClient(testLocalAddr)

				l, err := c.Acquire(lockName)
				if err != nil {
					b.Error(err)
					c.Close()
					return
				}
				err = l.Release()
				if err != nil {
					b.Error(err)
					c.Close()
					return
				}

				err = c.Close()
				if err != nil {
					b.Error(err)
					return
				}
			}
		})
	}
}
