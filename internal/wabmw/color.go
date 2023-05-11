package wabmw

import "github.com/aybabtme/rgbterm"

func red(text string) string {
	return rgbterm.FgString(text, 255, 0, 0)
}

func green(text string) string {
	return rgbterm.FgString(text, 0, 255, 0)
}

// func darkGreen(text string) string {
// 	return rgbterm.FgString(text, 69, 139, 0)
// }
//
// func darkCyan(text string) string {
// 	return rgbterm.FgString(text, 0, 139, 139)
// }

func cyan(text string) string {
	return rgbterm.FgString(text, 0, 255, 255)
}

func orange(text string) string {
	return rgbterm.FgString(text, 255, 165, 0)
}
