# distrilock
Distributed locking for pennies.

## levels of locking

1. the running instance tracks open files (one per lock) per connection
2. fcntl (collaborative) lock is acquired on the open file
3. fcntl (collaborative) lock is released on the open file, which is closed and forgotten

In case of errors at (2), file from (1) is closed.
Q: is state inconsistent in case of failures at (3)?

Tests must be added for all supported service scenarios.
A good test would be to peek locks being held consistently with internal maps states.

# Other links

https://github.com/karelzak/util-linux/blob/master/sys-utils/flock.c
http://www.tutorialspoint.com/unix_system_calls/fcntl.htm
http://www.informit.com/articles/article.aspx?p=23618&seqNum=4
http://www.grpc.io/docs/tutorials/basic/go.html#bidirectional-streaming-rpc-1
https://github.com/lomik/elock
https://github.com/komarov/switchman
https://redis.io/topics/distlock
https://martin.kleppmann.com/2016/02/08/how-to-do-distributed-locking.html
https://lwn.net/Articles/91268/
