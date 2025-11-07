package shellcode

import "workshop3_dev/internals/models"

type CommandShellcode interface {
	DoShellcode(dllBytes []byte, exportName string) (models.ShellcodeResult, error)
}
