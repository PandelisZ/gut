package provider

import (
	"context"
	"fmt"
	"sort"

	"gut/native/common"
	"gut/shared"
)

const libnutcoreDefaultElementInspectionLimit = 400

type libnutcoreElementInspectionProvider struct {
	client libnutcoreClient
}

type libnutcoreWindowElementNode struct {
	element  shared.WindowElement
	children map[int]*libnutcoreWindowElementNode
}

func NewLibnutcoreElementInspectionProvider(client libnutcoreClient) ElementInspectionProvider {
	return &libnutcoreElementInspectionProvider{client: client}
}

func (p *libnutcoreElementInspectionProvider) GetElements(ctx context.Context, windowHandle shared.WindowHandle, maxElements int) (shared.WindowElement, error) {
	if err := ctx.Err(); err != nil {
		return shared.WindowElement{}, err
	}
	if maxElements <= 0 {
		return shared.WindowElement{}, fmt.Errorf("%w: getElements [%s]", common.ErrInvalidToken, common.CapabilityAXElementSearch)
	}

	focusedWindow, err := p.focusedWindow(ctx, windowHandle)
	if err != nil {
		return shared.WindowElement{}, err
	}

	matches, err := p.client.SearchAXElements(common.AXElementSearchQuery{
		Scope:    common.AXSearchScopeFocusedWindow,
		Limit:    maxElements,
		MaxDepth: maxElements,
	})
	if err != nil {
		return shared.WindowElement{}, err
	}

	return buildLibnutcoreWindowElementTree(focusedWindow, matches), nil
}

func (p *libnutcoreElementInspectionProvider) FindElement(ctx context.Context, windowHandle shared.WindowHandle, description shared.WindowElementDescription) (shared.WindowElement, error) {
	elements, err := p.FindElements(ctx, windowHandle, description)
	if err != nil {
		return shared.WindowElement{}, err
	}
	if len(elements) == 0 {
		return shared.WindowElement{}, fmt.Errorf("element not found")
	}
	return elements[0], nil
}

func (p *libnutcoreElementInspectionProvider) FindElements(ctx context.Context, windowHandle shared.WindowHandle, description shared.WindowElementDescription) ([]shared.WindowElement, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !description.Valid() {
		return nil, fmt.Errorf("%w: window element description", common.ErrInvalidToken)
	}

	root, err := p.GetElements(ctx, windowHandle, libnutcoreDefaultElementInspectionLimit)
	if err != nil {
		return nil, err
	}

	matches := make([]shared.WindowElement, 0, 8)
	var visit func(shared.WindowElement)
	visit = func(element shared.WindowElement) {
		if windowElementMatchesDescription(element, description) {
			matches = append(matches, element)
		}
		for _, child := range element.Children {
			visit(child)
		}
	}
	visit(root)
	return matches, nil
}

func (p *libnutcoreElementInspectionProvider) focusedWindow(ctx context.Context, windowHandle shared.WindowHandle) (common.FocusedWindowMetadata, error) {
	focusedWindow, err := p.client.GetFocusedWindow()
	if err != nil {
		return common.FocusedWindowMetadata{}, err
	}
	if windowHandleFromNative(focusedWindow.Handle) == windowHandle {
		return focusedWindow, nil
	}

	if _, err := p.client.FocusWindow(windowHandleToNative(windowHandle)); err != nil {
		return common.FocusedWindowMetadata{}, err
	}
	if err := ctx.Err(); err != nil {
		return common.FocusedWindowMetadata{}, err
	}

	focusedWindow, err = p.client.GetFocusedWindow()
	if err != nil {
		return common.FocusedWindowMetadata{}, err
	}
	if focusedWindow.Handle != windowHandleToNative(windowHandle) {
		return common.FocusedWindowMetadata{}, fmt.Errorf("focused window handle %d does not match requested handle %d", focusedWindow.Handle, windowHandle)
	}
	return focusedWindow, nil
}

func buildLibnutcoreWindowElementTree(window common.FocusedWindowMetadata, matches []common.AXElementMatch) shared.WindowElement {
	root := &libnutcoreWindowElementNode{
		element: windowElementFromFocusedWindow(window),
	}
	for _, match := range matches {
		node := root
		for _, index := range match.Ref.Path {
			if node.children == nil {
				node.children = make(map[int]*libnutcoreWindowElementNode)
			}
			child, ok := node.children[index]
			if !ok {
				child = &libnutcoreWindowElementNode{}
				node.children[index] = child
			}
			node = child
		}
		node.element = mergeWindowElements(node.element, windowElementFromAXMatch(match))
	}
	root.element.Children = libnutcoreMaterializeWindowChildren(root.children)
	return root.element
}

func libnutcoreMaterializeWindowChildren(children map[int]*libnutcoreWindowElementNode) []shared.WindowElement {
	if len(children) == 0 {
		return nil
	}

	indexes := make([]int, 0, len(children))
	for index := range children {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)

	result := make([]shared.WindowElement, 0, len(indexes))
	for _, index := range indexes {
		child := children[index]
		element := child.element
		element.Children = libnutcoreMaterializeWindowChildren(child.children)
		result = append(result, element)
	}
	return result
}

func windowElementFromFocusedWindow(window common.FocusedWindowMetadata) shared.WindowElement {
	element := shared.WindowElement{
		Role:      optionalStringPointer(window.Role),
		SubRole:   optionalStringPointer(window.Subrole),
		Title:     optionalStringPointer(window.Title),
		IsFocused: boolPointer(window.Focused),
	}
	if window.RectKnown {
		region := regionFromNative(window.Rect)
		element.Region = &region
	}
	return element
}

func windowElementFromAXMatch(match common.AXElementMatch) shared.WindowElement {
	element := shared.WindowElement{
		Role:      optionalStringPointer(match.Metadata.Role),
		SubRole:   optionalStringPointer(match.Metadata.Subrole),
		Title:     optionalStringPointer(match.Metadata.Title),
		Value:     optionalStringPointer(match.Metadata.Value),
		IsEnabled: boolPointer(match.Metadata.Enabled),
		IsFocused: boolPointer(match.Metadata.Focused),
	}
	if match.Metadata.FrameKnown {
		region := regionFromNative(match.Metadata.Frame)
		element.Region = &region
	}
	return element
}

func mergeWindowElements(existing shared.WindowElement, incoming shared.WindowElement) shared.WindowElement {
	if incoming.Type != nil {
		existing.Type = incoming.Type
	}
	if incoming.Region != nil {
		existing.Region = incoming.Region
	}
	if incoming.Title != nil {
		existing.Title = incoming.Title
	}
	if incoming.Value != nil {
		existing.Value = incoming.Value
	}
	if incoming.IsFocused != nil {
		existing.IsFocused = incoming.IsFocused
	}
	if incoming.SelectedText != nil {
		existing.SelectedText = incoming.SelectedText
	}
	if incoming.IsEnabled != nil {
		existing.IsEnabled = incoming.IsEnabled
	}
	if incoming.Role != nil {
		existing.Role = incoming.Role
	}
	if incoming.SubRole != nil {
		existing.SubRole = incoming.SubRole
	}
	return existing
}

func windowElementMatchesDescription(element shared.WindowElement, description shared.WindowElementDescription) bool {
	if description.ID != "" {
		return false
	}
	if description.Role != "" && !windowElementStringEquals(element.Role, description.Role) {
		return false
	}
	if description.Type != "" && !windowElementStringEquals(element.Type, description.Type) {
		return false
	}
	if description.Title != nil && !windowElementStringMatches(element.Title, *description.Title) {
		return false
	}
	if description.Value != nil && !windowElementStringMatches(element.Value, *description.Value) {
		return false
	}
	if description.SelectedText != nil && !windowElementStringMatches(element.SelectedText, *description.SelectedText) {
		return false
	}
	return true
}

func windowElementStringEquals(value *string, expected string) bool {
	return value != nil && *value == expected
}

func windowElementStringMatches(value *string, matcher shared.StringMatcher) bool {
	return value != nil && matcher.Match(*value)
}

func optionalStringPointer(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func boolPointer(value bool) *bool {
	return &value
}
