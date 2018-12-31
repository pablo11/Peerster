package debug

import (
    "fmt"
)

func Debug(msg string) {
    fmt.Println("\033[0;33mDEBUG\033[0m: " + msg)
    fmt.Println()
}
