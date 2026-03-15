package shared

type MatchRequest[Needle any, ProviderData any] struct {
	Haystack     Image
	Needle       Needle
	Confidence   *float64
	ProviderData *ProviderData
}

type MatchResult[Location any] struct {
	Confidence float64
	Location   Location
	Error      error
}

func (r MatchResult[Location]) Succeeded() bool {
	return r.Error == nil
}
