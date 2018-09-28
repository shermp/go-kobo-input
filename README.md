# go-kobo-input
go-kobo-input provides those developing on Kobo ereaders basic touch handling capabilities.

## Installation & Usage
The sample application for go-kobo-input requires CGO to be enabled. A GCC ARM cross compiler is also required. The following environment variables need to be set as well:
```
GOOS=linux
GOARCH=arm
CGO_ENABLED=1

# Set C cross compiler variables
CC=/home/<user>/x-tools/arm-kobo-linux-gnueabihf/bin/arm-kobo-linux-gnueabihf-gcc
CXX=/home/<user>/x-tools/arm-kobo-linux-gnueabihf/bin/arm-kobo-linux-gnueabihf-g++
```

go-kobo-input can be obtained using go get:
```
go get github.com/shermp/go-kobo-input/...
```
Note that `go-fbink-v2` and `go-osk` is required for the sample application. However, `go get` should resolve this dependency.

Refer to `koboin-osk-sample/main.go` for a usage example. This sample also demonstrates usage of the `go-osk` onscreen keyboard.