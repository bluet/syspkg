# Go SysPkg: OSInfo

Go SysPkg is a library that provides system package management and operating system information. This documentation focuses on the `osinfo` package, which allows you to obtain information about the operating system.

## Installation

To install the `github.com/bluet/syspkg/osinfo` package, use the following command:

```bash
go get github.com/bluet/syspkg/osinfo
```

## Usage

To use the `osinfo` package, you need to import it into your Go program:

```go
import (
 "github.com/bluet/syspkg/osinfo"
)
```

### GetOSInfo

The primary function in the `osinfo` package is `GetOSInfo()`. It returns an `*OSInfo` struct, which contains information about the operating system, including:

- Name
- Distribution
- Version
- Architecture

Here's an example of how to use the `GetOSInfo()` function:

```go
package main

import (
 "fmt"
 "github.com/bluet/syspkg/osinfo"
)

func main() {
 osInfo, err := osinfo.GetOSInfo()
 if err != nil {
  fmt.Println("Error getting OS info:", err)
  return
 }

 fmt.Println("Name:", osInfo.Name)
 fmt.Println("Distribution:", osInfo.Distribution)
 fmt.Println("Version:", osInfo.Version)
 fmt.Println("Architecture:", osInfo.Arch)
}
```

In this example, the `GetOSInfo()` function is called, and the returned `*OSInfo` struct is used to print information about the operating system. If there is an error, it is handled and displayed to the user.

Run the example with the following command:

```bash
go run main.go
```

You should see output similar to the following:

```text
Name: linux
Distribution: Ubuntu
Version: 20.04
Architecture: amd64
```

The output will vary depending on the system you are running the program on.

## Summary

The `osinfo` package of the Go SysPkg library provides a simple and efficient way to obtain information about the operating system. The `GetOSInfo()` function is easy to use and returns a struct containing all the necessary information about the system's OS. This package can be useful for programs that need to detect the OS and perform tasks specific to a particular OS, distribution, or version.
