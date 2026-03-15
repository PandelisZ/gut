package shared

type SearchOptions[ProviderData any] struct {
	SearchRegion *Region
	Confidence   *float64
	ProviderData *ProviderData
}

func (o SearchOptions[ProviderData]) HasSearchRegion() bool {
	return o.SearchRegion != nil
}

func (o SearchOptions[ProviderData]) HasConfidence() bool {
	return o.Confidence != nil
}

func (o SearchOptions[ProviderData]) HasProviderData() bool {
	return o.ProviderData != nil
}
