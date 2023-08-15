# Why do we need a worker-pool in the first place?
Those who work in go must have come across this statement that go is cabable enough to run hundreds of thousands of go-routines at any given instance in time.
Apparently, this indeed is true. But then, in my humble opinion, we should raise a quiestion, why would we really want our service to run hundreds of thousands
of go-routines at any given instance in time.

Go's memory model is amazing and the effectiveness of go's memory management across multiple go-routines is equally good. As a matter of fact, go's GC is one of the fastest
ones amongst the GC oriented programming languages. As well, the small size of go-routines is very effective in the sense that it's far better when it comes to context
switching and swap-in and swap-out. But, no matter how effective the go-routine management is, running those many go-routines concurrently (I'm using concurrently and parallelly loosely
here, and thus, to mean the same) should still make us believe that we may need to revist the approach.

Keeping the concurrently running go-routines at a generously limited size is a better approach. This is important because creating too many worker go-routines can lead to performance
issues and resource contention. Though, size of each go-routine is small and typically is in the range of 2K to 4K, memory consumed by a huge number of concurrently running go-routines
is still going to be a costlier affair as, no matter how rich, the resources will still be limited.
The other important aspect is as the number of go-routines increases chances of thrashing will increase in equal proportions.
The sum effect of all this is the scalability issue. I'm dis-accounting the performance issue since it'll mostly be related to the job the work is going to execute.
As anyway since we spoke about performance issue and scalability issue, they're clearly different from each other:
Performance issue is, if our algorithim is taking more time for 1 unit of task execution, we've a performance problem.
Scalability issu is, if our system performs well for 1 unit of tasks execution, however, slows down if the size of tasks increases.

Thus, spinning off more and more go-routines to execute concurrently at any given instance in time
may result into sacalability issue.
So the first approach should be limiting the number of go-routines. But at the same time each job
needs to be executed. And this should happen judicially, meaning no job can be dropped in order to
control the number of concurrently running go-routines. So, the only possibility remains is to
control the number of concurrently running go-routines.
The best way is through creating a team of go-routines with a fixed number of team members.
Thus, each go-routine in the team may either be free or be executing a job at any given instance
in time. This team is go-routine pool, also termed as worker-pool in generic terms.

Implementation strategies:
Buffered channels for jobs and worker-pool.

There're 2 ends, as we know of, the publishers who publish the data and the subscribers who've had subscribed to the relevant 
data. Worker-pool is the implementation strategy for publisher-subscriber model. 
Jobs queue is the datastructure that sits between publisher and subscriber. A publisher adds data onto the job queue. Subscriber is
waiting for some data to arrive in the job queue. Accordingly, a

Job queue is a buffered channel.
