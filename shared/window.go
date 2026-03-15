package shared

import "regexp"

type WindowHandle uint64

type StringMatcher struct {
	Exact   string
	Pattern *regexp.Regexp
}

func MatchString(exact string) StringMatcher {
	return StringMatcher{Exact: exact}
}

func MatchPattern(pattern *regexp.Regexp) StringMatcher {
	return StringMatcher{Pattern: pattern}
}

func (m StringMatcher) Match(value string) bool {
	if m.Pattern != nil {
		return m.Pattern.MatchString(value)
	}
	return value == m.Exact
}

func (m StringMatcher) Valid() bool {
	return m.Pattern != nil || m.Exact != ""
}

type WindowElement struct {
	Type         *string
	Region       *Region
	Title        *string
	Value        *string
	IsFocused    *bool
	SelectedText *string
	IsEnabled    *bool
	Role         *string
	SubRole      *string
	Children     []WindowElement
}

type WindowElementDescription struct {
	ID           string
	Role         string
	Type         string
	Title        *StringMatcher
	Value        *StringMatcher
	SelectedText *StringMatcher
}

func (d WindowElementDescription) Valid() bool {
	return d.ID != "" || d.Role != "" || d.Type != "" ||
		(d.Title != nil && d.Title.Valid()) ||
		(d.Value != nil && d.Value.Valid()) ||
		(d.SelectedText != nil && d.SelectedText.Valid())
}
