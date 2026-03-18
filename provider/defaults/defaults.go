package defaults

import (
	"github.com/PandelisZ/gut/clipboard"
	"github.com/PandelisZ/gut/imageio"
	"github.com/PandelisZ/gut/imageproc"
	"github.com/PandelisZ/gut/provider"
)

func RegisterAuxiliary(registry *provider.Registry) {
	if registry == nil {
		return
	}

	processor := imageproc.NewProcessor()
	registry.RegisterClipboard(clipboard.NewSystemProvider())
	registry.RegisterImageReader(imageio.NewReader())
	registry.RegisterImageWriter(imageio.NewWriter())
	registry.RegisterImageProcessor(processor)
	registry.RegisterColorFinder(imageproc.NewColorFinder(processor))
	registry.RegisterImageFinder(imageproc.NewUnavailableImageFinder())
	registry.RegisterTextFinder(imageproc.NewUnavailableTextFinder())

	windowProvider, err := registry.Window()
	if err != nil {
		registry.RegisterWindowFinder(imageproc.NewUnavailableWindowFinder())
		return
	}
	registry.RegisterWindowFinder(imageproc.NewWindowFinder(windowProvider))
}
