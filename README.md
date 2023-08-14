# Why do we need a worker-pool in the first place?
Those who work in go must have come across this statement that go is cabable enough to run hundreds of thousands of go-routines at any given instance in time.
Apparently, this indeed is true. But then, in my humble opinion, we should raise a quiestion, why would we really want our service to run hundreds of thousands
of go-routines at any given instance in time.

Go's memory model is amazing and the effectiveness of go's memory management across multiple go-routines is equally good. As a matter of fact, go's GC is one of the fastest
ones amongst the GC oriented programming languages. As well, the small size of go-routines is very effective in the sense that it's far better when it comes to context
switching and swap-in and swap-out. But, no matter how effective the go-routine management is, running those many go-routines concurrently (I'm using concurrently and parallelly loosely
here, and thus, to mean the same) should still make us believe that we may need to revist the approach.

Keeping the concurrently running go-routines at a generously limited size is a better approach. This is important because creating too many worker go-routines can lead to performance issues and resource contention.
