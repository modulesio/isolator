// +build !linux

package linuxsandbox

import "github.com/modulesio/butler/mansion"

func Register(ctx *mansion.Context) {
	// don't register anything
	return
}
