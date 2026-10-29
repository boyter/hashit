// SPDX-License-Identifier: MIT

//go:build !solaris

package processor

import (
	"os"

	"github.com/gosuri/uiprogress"
)

// progressSupported is false on platforms where the progress bars cannot be built
const progressSupported = true

// progressBar is a single --progress bar, set between 0 and UiBarMax
type progressBar interface {
	Set(n int) error
}

// progressBars renders the --progress bars
var progressBars *uiprogress.Progress

// startProgress starts rendering the --progress bars, to standard error so the bars do not end
// up in redirected output
func startProgress() {
	progressBars = uiprogress.New()
	progressBars.SetOut(os.Stderr)
	progressBars.Start()
}

// newProgressBar adds a bar which displays label after it
func newProgressBar(label func() string) progressBar {
	bar := progressBars.AddBar(UiBarMax)
	bar.AppendFunc(func(b *uiprogress.Bar) string {
		return label()
	})
	return bar
}
