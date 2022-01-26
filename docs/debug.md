#  Debugging a Terraform provider 

## Understanding the Design

In order to do debug, you first have to understand how Go builds apps, and then how Terraform works with it.

Every Terraform provider is a sort of `module`. In order to support an open, modular system, in almost any language, you need to be able to dynamically load modules and interact with them. Terraform is no exception.

However, the Golang team long ago decided to compile to statically linked applications; any dependencies you have will be compiled into a single binary. Unlike in other native languages (like C, or C++), a `.dll` or `.so` is not used; there is no dynamic library to load at runtime and thus, modularity becomes a whole other trick. This is done to avoid the notorious **dll hell** that was so common up until most modern systems included some kind of dependency management. And yes, it can still be an issue.

Every Terraform provider is its own mini gRPC server. When the Terraform client runs your provider, it actually starts a new process that is your provider, and connects to it through this gRPC channel. In other words, the `terraform` binary is the gRPC _client_, and the Terraform provider is the gRPC _server_. Compounding the problem is that the lifetime of your provider process is ephemeral, potentially lasting no more than a few seconds. It's this process where you need to connect with your debugger.

### Normal Debugging

Normally, you would directly spin-up your app, and it would load modules into application memory. That's why you can actually debug it, because your debugger knows how to find the exact memory address for your provider. However, you don't have this arrangement, and you need to do a _remote_ debug session.

### The Conundrum

So, you don't load Terraform directly, and even if you did, your provider (e.g., gRPC server) is in the memory space of an entirely different process; and that lasts no more than a few seconds.

## The Solution

1. You need the debugging tool [delve](https://github.com/go-delve/delve).

1. You are going to have to place a little bit of shim code close to the spot in the code where you want to begin debugging. We need to stop this provider process from exiting before we can connect. So, put this bit of code in place:

    ```go
    connected := false
    for !connected {
        time.Sleep(time.Second) // set breakpoint here
    }
    ```

    This code creates an infinite sleep loop, but it's essential to solving the problem.

1. Place a break point right inside this loop. It won't do anything, yet.

1. Now run the Terraform commands you need to, to engage the code you're desiring to debug. Upon doing so, Terraform will stop as it waits on a response from your provider.

1. You must now tell `delve` to connect to this remote process using it's PID. This isn't as hard as it seems. Run these commands:

    ```bash
    dlv \
        --listen=:2345 \
        --headless=true \
        --api-version=2 \
        --accept-multiclient attach \
        $(pgrep terraform-provider-artifactory)
    ```

    The last argument gets the `PID` for your provider and supplies it to `delve` to connect. Immediately upon running this command, you're going to hit your break point. Please make sure to substitute `terraform-provider-artifactory` for your provider name.

1. To exit this infinite loop, use your debugger to set `connected` to `true`. By doing so you change the loop predicate 
and it will exit this loop on the next iteration.

1. _DEBUG!_ - At this point you can, step, watch, drop the call stack, etc. Your entire arsenal is available.
