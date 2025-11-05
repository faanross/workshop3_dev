package shellcode

import (
	"fmt"
	"workshop3_dev/internals/models"
)

// macShellcode implements the CommandShellcode interface for Darwin/MacOS.
type macShellcode struct{}

// New is the constructor for our Mac-specific Shellcode command
func New() CommandShellcode {
	return &macShellcode{}
}

func (ms *macShellcode) DoShellcode(shellcode []byte, exportName string) (models.LoadResult, error) {
	fmt.Println("|❗ SHELLCODE DOER MACOS| This feature has not yet been implemented for MacOS.")

	result := models.LoadResult{
		Message: "FAILURE",
	}
	return result, nil
}
