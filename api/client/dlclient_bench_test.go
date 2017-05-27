package client_test

import (
	"testing"
)

func BenchmarkLocksTaking(b *testing.B) {
	type lockTakingTest struct {
		name    string
		nameGen func() string
		cs      *clientSuite
	}

	fixedLockName := generateLockName(b)

	benchmarks := []lockTakingTest{
		{
			name: "TCP random locks",
			nameGen: func() string {
				return generateLockName(b)
			},
			cs: tcpClientSuite,
		},
		{
			name: "TCP fixed locks",
			nameGen: func() string {
				return fixedLockName
			},
			cs: tcpClientSuite,
		},
		{
			name: "websocket random locks",
			nameGen: func() string {
				return generateLockName(b)
			},
			cs: websocketClientSuite,
		},
		{
			name: "websocket fixed locks",
			nameGen: func() string {
				return fixedLockName
			},
			cs: websocketClientSuite,
		},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				lockName := bm.nameGen()

				c := bm.cs.createLocalClient()

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
