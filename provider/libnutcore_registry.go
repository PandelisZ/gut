package provider

func RegisterLibnutcoreProviders(registry *Registry, client libnutcoreClient) {
	registry.RegisterKeyboard(NewLibnutcoreKeyboardProvider(client))
	registry.RegisterMouse(NewLibnutcoreMouseProvider(client))
	registry.RegisterScreen(NewLibnutcoreScreenProvider(client))
	registry.RegisterWindow(NewLibnutcoreWindowProvider(client))
}
