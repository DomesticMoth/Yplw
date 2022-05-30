# confer
Yet another configuration library for go  

There is a common program configuration case in which the program receives the path to the file as the only argument and reads the configuration from this file.  
This library provides the only function that parses the toml file along the path passed in the arguments.  

# Install
```
$ go get github.com/DomesticMoth/confer
```

# Usage
```go
package main

import  (
	"github.com/DomesticMoth/confer"
	"fmt"
)

type Config struct {
	SomeParam string
}

func main() {
  var conf Config
  err := confer.LoadConfig([]string{}, &conf)
  if err != nil {
    panic(err)
  }
  fmt.Println(conf)
}
```

It also provides the ability to specify one or more default paths by which the file will be searched if no path was passed through arguments or if the configuration could not be loaded using the passed path.  

```go
package main

import  (
	"github.com/DomesticMoth/confer"
	"fmt"
)

type Config struct {
	SomeParam string
}

func main() {
  defaultLocalPath := "~/cofigfile"
  defaultGloabalPath := "/etc/cofigfile"
  var conf Config
  err := confer.LoadConfig([]string{defaultLocalPath, defaultGloabalPath}, &conf)
  if err != nil {
    panic(err)
  }
  fmt.Println(conf)
}
```
