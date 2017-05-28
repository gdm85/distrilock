	// resolve daemon address
	addr, err := net.ResolveTCPAddr("tcp", ":13123")
	if err != nil {
		panic(err)
	}
	
	// create client
	c := tcp.New(addr, time.Second*3, time.Second*3, time.Second*3)
	
	// acquire lock
	l, err := c.Acquire("my-named-lock")
	if err != nil {
		panic(err)
	}
	
	// initialise a goroutine verifying lock at a specific interval
	brokenLock := make(chan error)
	done := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		t := time.NewTicker(time.Millisecond * 400)
		defer t.Stop()
		defer func() { finished <- struct{}{} }
		tickChan := t.C
		for {
			select {
				case <-done:
					close(brokenLock)
					return
				case <-tickChan:
					err := l.Verify()
					if err != nil {
						brokenLock <- err
						close(brokenLock)
						return
					}
			}
		}
     }()
     
     // start doing some intensive work
     for !completed {
		select {
			case err := <-brokenLock:
				// lock was broken, operations cannot continue
				panic(err)
			default:
				// nothing to pick from channel, fast exit
		}

		///
		/// ... do some heavy work here, then iterate for some more heavy work
		///
     }
     
	// stop goroutine
	done <- struct{}{}
	close(done)
	<-finished
	
	// release lock
	err = l.Release()
	if err != nil {
		panic(err)
	}
	
	// close connection
	err = c.Close()
	if err != nil {
		panic(err)
	}
