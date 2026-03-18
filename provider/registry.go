package provider

import (
	"fmt"

	gutlog "github.com/PandelisZ/gut/log"
)

type MissingProviderError struct {
	Name string
}

func (e MissingProviderError) Error() string {
	return fmt.Sprintf("provider %q is not registered", e.Name)
}

func (e MissingProviderError) Is(target error) bool {
	switch t := target.(type) {
	case MissingProviderError:
		return e.Name == t.Name
	case *MissingProviderError:
		return t != nil && e.Name == t.Name
	default:
		return false
	}
}

var (
	ErrMissingKeyboardProvider          = MissingProviderError{Name: "keyboard"}
	ErrMissingMouseProvider             = MissingProviderError{Name: "mouse"}
	ErrMissingScreenProvider            = MissingProviderError{Name: "screen"}
	ErrMissingWindowProvider            = MissingProviderError{Name: "window"}
	ErrMissingAccessibilityProvider     = MissingProviderError{Name: "accessibility"}
	ErrMissingImageFinderProvider       = MissingProviderError{Name: "image-finder"}
	ErrMissingImageReaderProvider       = MissingProviderError{Name: "image-reader"}
	ErrMissingImageWriterProvider       = MissingProviderError{Name: "image-writer"}
	ErrMissingImageProcessorProvider    = MissingProviderError{Name: "image-processor"}
	ErrMissingTextFinderProvider        = MissingProviderError{Name: "text-finder"}
	ErrMissingWindowFinderProvider      = MissingProviderError{Name: "window-finder"}
	ErrMissingColorFinderProvider       = MissingProviderError{Name: "color-finder"}
	ErrMissingElementInspectionProvider = MissingProviderError{Name: "element-inspection"}
	ErrMissingClipboardProvider         = MissingProviderError{Name: "clipboard"}
)

type Registry struct {
	keyboard          KeyboardProvider
	mouse             MouseProvider
	screen            ScreenProvider
	window            WindowProvider
	accessibility     AccessibilityProvider
	logger            gutlog.Logger
	imageFinder       ImageFinder
	imageReader       ImageReader
	imageWriter       ImageWriter
	imageProcessor    ImageProcessor
	textFinder        TextFinder
	windowFinder      WindowFinder
	colorFinder       ColorFinder
	elementInspection ElementInspectionProvider
	clipboard         ClipboardProvider
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) RegisterKeyboard(provider KeyboardProvider)           { r.keyboard = provider }
func (r *Registry) RegisterMouse(provider MouseProvider)                 { r.mouse = provider }
func (r *Registry) RegisterScreen(provider ScreenProvider)               { r.screen = provider }
func (r *Registry) RegisterWindow(provider WindowProvider)               { r.window = provider }
func (r *Registry) RegisterAccessibility(provider AccessibilityProvider) { r.accessibility = provider }
func (r *Registry) RegisterLogger(provider gutlog.Logger)                { r.logger = provider }
func (r *Registry) RegisterImageFinder(provider ImageFinder)             { r.imageFinder = provider }
func (r *Registry) RegisterImageReader(provider ImageReader)             { r.imageReader = provider }
func (r *Registry) RegisterImageWriter(provider ImageWriter)             { r.imageWriter = provider }
func (r *Registry) RegisterImageProcessor(provider ImageProcessor)       { r.imageProcessor = provider }
func (r *Registry) RegisterTextFinder(provider TextFinder)               { r.textFinder = provider }
func (r *Registry) RegisterWindowFinder(provider WindowFinder)           { r.windowFinder = provider }
func (r *Registry) RegisterColorFinder(provider ColorFinder)             { r.colorFinder = provider }
func (r *Registry) RegisterElementInspection(provider ElementInspectionProvider) {
	r.elementInspection = provider
}
func (r *Registry) RegisterClipboard(provider ClipboardProvider) { r.clipboard = provider }

func (r *Registry) Keyboard() (KeyboardProvider, error) {
	if r.keyboard == nil {
		return nil, ErrMissingKeyboardProvider
	}
	return r.keyboard, nil
}

func (r *Registry) Mouse() (MouseProvider, error) {
	if r.mouse == nil {
		return nil, ErrMissingMouseProvider
	}
	return r.mouse, nil
}

func (r *Registry) Screen() (ScreenProvider, error) {
	if r.screen == nil {
		return nil, ErrMissingScreenProvider
	}
	return r.screen, nil
}

func (r *Registry) Window() (WindowProvider, error) {
	if r.window == nil {
		return nil, ErrMissingWindowProvider
	}
	return r.window, nil
}

func (r *Registry) Accessibility() (AccessibilityProvider, error) {
	if r.accessibility == nil {
		return nil, ErrMissingAccessibilityProvider
	}
	return r.accessibility, nil
}

func (r *Registry) Logger() gutlog.Logger {
	if r.logger == nil {
		return gutlog.Nop()
	}
	return r.logger
}

func (r *Registry) ImageFinder() (ImageFinder, error) {
	if r.imageFinder == nil {
		return nil, ErrMissingImageFinderProvider
	}
	return r.imageFinder, nil
}

func (r *Registry) ImageReader() (ImageReader, error) {
	if r.imageReader == nil {
		return nil, ErrMissingImageReaderProvider
	}
	return r.imageReader, nil
}

func (r *Registry) ImageWriter() (ImageWriter, error) {
	if r.imageWriter == nil {
		return nil, ErrMissingImageWriterProvider
	}
	return r.imageWriter, nil
}

func (r *Registry) ImageProcessor() (ImageProcessor, error) {
	if r.imageProcessor == nil {
		return nil, ErrMissingImageProcessorProvider
	}
	return r.imageProcessor, nil
}

func (r *Registry) TextFinder() (TextFinder, error) {
	if r.textFinder == nil {
		return nil, ErrMissingTextFinderProvider
	}
	return r.textFinder, nil
}

func (r *Registry) WindowFinder() (WindowFinder, error) {
	if r.windowFinder == nil {
		return nil, ErrMissingWindowFinderProvider
	}
	return r.windowFinder, nil
}

func (r *Registry) ColorFinder() (ColorFinder, error) {
	if r.colorFinder == nil {
		return nil, ErrMissingColorFinderProvider
	}
	return r.colorFinder, nil
}

func (r *Registry) ElementInspection() (ElementInspectionProvider, error) {
	if r.elementInspection == nil {
		return nil, ErrMissingElementInspectionProvider
	}
	return r.elementInspection, nil
}

func (r *Registry) Clipboard() (ClipboardProvider, error) {
	if r.clipboard == nil {
		return nil, ErrMissingClipboardProvider
	}
	return r.clipboard, nil
}
