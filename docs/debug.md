#  Debugging a TerraForm provider 

## Understanding the design

In order to do it, you first have to understand how Go builds apps, and then how terraform works with it.

Every terraform provider is a sort of `module`. In order to support an open, modular system, in almost any language, you need to be able to dynamically load modules and interact with them. Terraform is no exception.

However, the go lang team long ago decided to compile to statically linked applications; 
any dependencies you have will be compiled into 1 single binary. Unlike in other native languages (like C, or C++), a 
`.dll` or `.so` is not used; there is no dynamic library to load at runtime and thus, modularity becomes a whole other trick.
This is done to avoid the notorious **dll hell** that was so common up until most modern systems included some 
kind of dependency management. And yes, it can still be an issue.

Every terraform provider is its own mini RPC server. When terraform runs your provider, it actually starts a new process that is your provider, and connects to it through 
this RPC channel. Compounding the problem is that the lifetime of your provider process is very much 
ephemeral; potentially lasting no more and a few seconds. It's this process you need to connect to with your debugger

### Normal debugging
Normally, you would directly spin-up your app, and it would load modules into application memory. That's why you can actually
debug it, because your debugger knows how to find the exact memory address for your provider. However, you don't have 
this arrangement, and you need to do a _remote_ debug session. 

### The conundrum
So, you don't load terraform directly, and even if you did, your `module` (a.k.a your provider) is in the memory 
space of an entirely different process; and that lasts no more than a few seconds, potentially.  

## The solution

1. You need the debugging tool [delve](https://github.com/go-delve/delve).
2. You are going to have to place a little bit of shim code close to the spot in the code where you want to begin
debugging. We need to stop this provider process from exiting before we can connect. So, put this bit of code in place:
```go
	connected := false
	for !connected {
		time.Sleep(time.Second) // set breakpoint here
	}
```
This code effectively creates an infinite sleep loop; but that's actually essential to solving the problem.

3. Place a break point right inside this loop. It won't do anything, yet. 
4. Now run the terraform commands you need to, to engage the code you're desiring to debug. Upon doing so,
terraform will basically stop, as it waits on a response from you provider; because you put an infinite sleep loop in
5. You must now tell `delve` to connect to this remote process using it's PID. This isn't as hard as it seems.
Run this commands:
`dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient attach $(pgrep terraform-provider-artifactory)`
The last argument gets the `PID` for your provider and supplies it to `delve` to connect. Immediately upon running this 
command, you're going to hit your break point. Please make sure to substitute `terraform-provider-artifactory` for your provider name
6. To exit this infinite loop, use your debugger to set `connected` to `true`. By doing so you change the loop predicate 
and it will exit this loop on the next iteration.
7. *DEBUG!* - At this point you can, step, watch, drop the call stack, etc. Your whole arsenal is available.
