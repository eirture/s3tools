package build

import (
	"fmt"
)

var (
	Version = "unknown"
	Date    string
	Release string
)

func PrintVersion() {
	fmt.Printf("s3tools version %s (%s)\n", Version, Date)
	if Release != "" {
		fmt.Println(Release)
	}
}
