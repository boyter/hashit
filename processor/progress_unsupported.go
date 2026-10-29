// SPDX-License-Identifier: MIT

// the solaris constraint also applies to illumos
//go:build solaris

package processor

// progressSupported is false as the terminal library used for the progress bars does not build
// on illumos or solaris
const progressSupported = false

// progressBar is a single --progress bar, set between 0 and UiBarMax
type progressBar interface {
	Set(n int) error
}

func startProgress() {}

func newProgressBar(label func() string) progressBar {
	return nil
}
