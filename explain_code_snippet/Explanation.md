# Problem 3: Go Concurrency Analysis

## The Code

```go
package main
import "fmt"

func main() {
    cnp := make(chan func(), 10)
    
    for i := 0; i < 4; i++ {
        go func() {
            for f := range cnp {
                f()
            }
        }()
    }
    
    cnp <- func() {
        fmt.Println("HERE1")
    }
    fmt.Println("Hello")
}
```

Output: `Hello` (that's it - "HERE1" never prints)

---

## Q1: What are the key constructs?

- `make(chan func(), 10)` - Creates a buffered channel that holds functions. The buffer size of 10 means you can send 10 functions without blocking. Think of it like a queue with 10 slots.

- `go func() { ... }()` - Spawns a new goroutine. The function runs concurrently in the background.

- `for f := range cnp`** - Continuously pulls functions from the channel and blocks when empty. This only stops if the channel gets closed (which never happens here).

- `cnp <- func() { ... }`** - Sends a function into the channel. Since it's buffered with space available, this doesn't block.

---

## Q2: Where would you use these patterns?

The worker pool pattern (what this code implements) is common for:

- Background job processing - Handle API requests, send emails, process images
- Rate limiting - Control how many operations run simultaneously
- Pipeline architectures - Chain processing stages together
- Dynamic task execution - Execute different behaviors without tight coupling

Example use case:
```go
// Image thumbnail generator
imageQueue := make(chan Image, 50)
for i := 0; i < 4; i++ {
    go func() {
        for img := range imageQueue {
            generateThumbnail(img)
        }
    }()
}
```

Channels of functions specifically let you pass behavior around, useful for callback systems or command patterns.

---

## Q3: Why 4 iterations?

Creates 4 worker goroutines - a basic worker pool. Each worker waits on the same channel, so when you send a task, one of them picks it up.

Why 4? Usually matches CPU cores for CPU-bound work. More workers = more concurrency but also more scheduling overhead. For I/O work you might use more, for CPU work you'd typically match core count.

```
            ┌─ Worker 1: waiting on channel
            ├─ Worker 2: waiting on channel  
Channel ────┼─ Worker 3: waiting on channel
            └─ Worker 4: waiting on channel
```

When a function arrives, whichever worker is available grabs it.

---

## Q4: Why buffer size 10?

The difference between buffered and unbuffered channels:

**Unbuffered (size 0):**
- Send blocks until someone receives
- Forces strict synchronization
- Like a handoff - both parties must be ready

**Buffered (size 10):**
- Send only blocks when full
- Decouples sender/receiver timing
- Handles burst traffic better

With buffer=10, you can rapidly queue 10 functions before blocking. Good for smoothing out speed differences between producers and consumers.

Trade-offs:
- Too small: frequent blocking, poor throughput
- Too large: memory waste, hides backpressure issues
- 10-100 is typical for moderate workloads

---

## Q5: Why doesn't "HERE1" print?

Race condition. Here's what happens:

```
Time    Main goroutine              Worker goroutines
----    ---------------             -----------------
t0      Create channel
t1      Spawn 4 workers             → All waiting on channel
t2      Send function               → One worker receives it
t3      Print "Hello"               → Worker starting to run function
t4      main() exits                → ☠️ ALL GOROUTINES KILLED
        Program terminates          
```

When `main()` returns, the entire program exits immediately. Doesn't matter if other goroutines are still running - they get terminated. The worker grabbed the function but didn't get time to execute `fmt.Println("HERE1")` before everything shut down.

### Fix #1: Wait for completion

```go
func main() {
    cnp := make(chan func(), 10)
    done := make(chan bool)
    
    for i := 0; i < 4; i++ {
        go func() {
            for f := range cnp {
                f()
                done <- true
            }
        }()
    }
    
    cnp <- func() {
        fmt.Println("HERE1")
    }
    
    <-done  // Wait for worker to finish
    fmt.Println("Hello")
}
```

### Fix #2: Use sync.WaitGroup (better for multiple tasks)

```go
func main() {
    cnp := make(chan func(), 10)
    var wg sync.WaitGroup
    
    for i := 0; i < 4; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for f := range cnp {
                f()
            }
        }()
    }
    
    cnp <- func() {
        fmt.Println("HERE1")
    }
    
    close(cnp)
    wg.Wait()
    fmt.Println("Hello")
}
```

Note the `close(cnp)` - without it, workers stay stuck in the range loop forever. Closing signals "no more data coming" so the range loop can exit.

### Fix #3: Quick hack (not recommended)

```go
time.Sleep(100 * time.Millisecond)  // Just wait a bit
fmt.Println("Hello")
```

Works but fragile - you're guessing how long the work takes.

---

## Summary

This code implements a worker pool but has a fatal flaw: main exits before workers finish. The goroutines are killed mid-execution. You need explicit synchronization (WaitGroup, channels, or similar) to coordinate lifecycle between main and background workers.

The pattern itself is solid and widely used in production - just needs proper cleanup handling.