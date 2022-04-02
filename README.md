# hw7_microservice

In this task, you will learn how to build a microservice based on the grpc framework.

*In this task, you cannot use global variables, store what you need in the fields of the structure that lives in the closure*

You will need to implement:

* Generate required code from proto file
* Microservice base with the ability to stop the server
* ACL - access control from different clients
* System of logging of called methods
* A system for collecting assembly statistics (just counters) on called methods

Microservice will consist of 2 parts:
* Some business logic. In our example, it does nothing, it is enough just to call it
* Administration module, where logging and statistics are located

With the first, everything is simple, there is no logic there.

The second one is more interesting. As a rule, in real microservices, both logging and statistics work in a single instance, but in our case they will be available via the streaming interface to those who connect to the service. This means that 2 logging clients can connect to the service and both will receive a stream of logs. Also, 2 (or more) statistics modules can connect to the service with different intervals for receiving statistics (for example, every 2, 3 and 5 seconds) and it will be sent asynchronously over each interface.

Since asynchronous has already been mentioned, the task will have goroutines, timers, mutexes, a context with timeouts / completion.

Task features:

The contents of service.pb.go (which you got when generating the proto file) you need to put in service.go to load 1 file
You cannot use global variables in this job. All that we need - store in the fields of the structure.

Run tests with go test -v -race
