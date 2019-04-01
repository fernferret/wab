// +build dev

package wab

import "net/http"

var WebUI http.FileSystem = http.Dir("webapp/build")
