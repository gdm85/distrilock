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

## FILE LOCKING DEMO

(from http://voyager.deanza.edu/~perry/lock.html)

### Write Lock Setter
```c
#include <sys/types.h>
#include <unistd.h>      
#include <fcntl.h>
main()
{
  int fd;
  struct flock lock, savelock;

  fd = open("book.dat", O_RDWR);
  lock.l_type    = F_WRLCK;   /* Test for any lock on any part of file. */
  lock.l_start   = 0;
  lock.l_whence  = SEEK_SET;
  lock.l_len     = 0;        
  savelock = lock;
  fcntl(fd, F_GETLK, &lock);  /* Overwrites lock structure with preventors. */
  if (lock.l_type == F_WRLCK)
  {
     printf("Process %ld has a write lock already!\n", lock.l_pid);
     exit(1);
  }
  else if (lock.l_type == F_RDLCK)
  {
     printf("Process %ld has a read lock already!\n", lock.l_pid);
     exit(1);
  }
  else
     fcntl(fd, F_SETLK, &savelock);
  pause();
}
```
### Read Lock Setter

```c
#include <stdio.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <unistd.h>
main()
{
  struct flock lock, savelock;
  int fd;
   
  fd = open("book.dat", O_RDONLY);
  lock.l_type = F_RDLCK;
  lock.l_start = 0;
  lock.l_whence = SEEK_SET;
  lock.l_len = 50;
  savelock = lock;
  fcntl(fd, F_GETLK, &lock);
  if (lock.l_type == F_WRLCK)
  {
      printf("File is write-locked by process %ld.\n", lock.l_pid);
      exit(1);
  }
  fcntl(fd, F_SETLK, &savelock);
  pause();
}
```
## SAMPLE PROGRAM EXECUTIONS
```bash
$ wl &
[1]	20866
$ rl
File is write-locked by process 20866.

$ rl &
[1]	20868
$ wl
Process 20868 has a read lock already!
```

##  Critical Points About File Locking

1.  A write lock on any region prevents ANY kind of lock on even a single
    byte of that region.   

2.  A read lock on any region prevents a write lock on even a single byte
    of that region.

3.  Fcntl with F_GETLK or F_SETLK return -1 on an "error".  However, an
    error with F_GETLK means that it cannot obtain the desired lock
    information.  With F_SETLK it means that the desired lock cannot be
    obtained.

3.  When using fcntl with F_GETLK, the l_type component of the flock
    struct is overwritten by any existing lock which would prevent the
    lock whose attributes are in the struct flock * third argument.
    If nothing will prevent the lock given in this struct, the l_type
    component will have the value F_UNLCK.

4.  If the value of the l_len component of a struct flock is 0 then the
    length of the region to be tested/locked is the rest of the file
    starting at the value given in the l_start component.

5.  The l_start component's meaning is relative to the l_whence component
    of the flock struct.  The l_whence component can be SEEK_SET (beginning
    of file), SEEK_CUR (current position of file pointer), or SEEK_END
    (end of file).  Thus, l_start must be zero or positive if l_whence is
    SEEK_SET and zero or negative if l_whence is SEEK_END.

6.  Read locks do not prevent other read locks on the same region.  

7.  You must remember to F_UNLCK a region after your need for a read or
    write lock has expired.  

8.  With descriptors, there is no fseek/ftell.  There is only lseek which
    returns the same value as ftell and whose prototype is:

          #include <sys/types.h>
          #include <unistd.h> 
          off_t lseek(int fd, off_t offset, int whence);

     Lseek's return value is the offset after the move of the file pointer.
     Lseek(fd, 0, SEEK_CUR) is like ftell.

9.   If you want to create a lock, do NOT call fcntl with F_GETLK first.
     Just try for the lock with F_SETLK and if fcntl returns a negative
     value then you couldn't get the lock.  F_GETLK is an "info only please"
     request.
