## Udf filesystem golang library

- Non-optimized
- Some features may be broken
- `recovery()` style error handling interface
- Tested only with certain ISOs (e.g. Windows ISOs)

It's all because I has reached required functionality for me.

## Example

```go
package main

import (
	"fmt"
	"os"
	"github.com/saidelike/udf"
)

func main() {
	rdr, _ := os.Open("example.iso")
	u, _ := udf.NewUdfFromReader(r)
	for _, f := range u.ReadDir(nil) {
		fmt.Printf("%s %-10d %-20s %v\n", f.Mode().String(), f.Size(), f.Name(), f.ModTime())
	}
}

```
Output:
```
-r-xr-xr-x 57         system.cnf           2006-02-11 00:00:00 +0000 UTC
-r-xr-xr-x 1911580    SCUS_973.99          2006-03-15 00:00:00 +0000 UTC
-r-xr-xr-x 278305     ioprp300.img         2005-11-14 00:00:00 +0000 UTC
-r-xr-xr-x 6641       sio2man.irx          2005-10-18 00:00:00 +0000 UTC
-r-xr-xr-x 15653      dbcman.irx           2005-10-18 00:00:00 +0000 UTC
```

See [isoinfo.go](isoinfo/isoinfo.go) for complete example.

## Specification

* http://www.osta.org/specs/pdf/udf260.pdf
