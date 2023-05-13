package proto

import (
	_ "embed"
)

// This module only serves to hod the embedded ui
// This is because go's embed refuses (for security reasons) to embed
// relative paths (which is fine). So I moved the ui.go over here!

//go:embed greeter.proto
var Greeter string
