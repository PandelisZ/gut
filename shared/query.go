package shared

type QueryType string

const (
	QueryTypeText          QueryType = "text"
	QueryTypeWindow        QueryType = "window"
	QueryTypeColor         QueryType = "color"
	QueryTypeWindowElement QueryType = "window-element"
)

type TextQueryBy struct {
	Word string
	Line string
}

type TextQuery struct {
	ID string
	By TextQueryBy
}

func NewWordQuery(id, word string) TextQuery {
	return TextQuery{ID: id, By: TextQueryBy{Word: word}}
}

func NewLineQuery(id, line string) TextQuery {
	return TextQuery{ID: id, By: TextQueryBy{Line: line}}
}

func (q TextQuery) Type() QueryType {
	return QueryTypeText
}

func (q TextQuery) IsWord() bool {
	return q.By.Word != "" && q.By.Line == ""
}

func (q TextQuery) IsLine() bool {
	return q.By.Line != "" && q.By.Word == ""
}

func (q TextQuery) Valid() bool {
	return q.IsWord() || q.IsLine()
}

type WindowQueryBy struct {
	Title StringMatcher
}

type WindowQuery struct {
	ID string
	By WindowQueryBy
}

func NewWindowQuery(id string, title StringMatcher) WindowQuery {
	return WindowQuery{ID: id, By: WindowQueryBy{Title: title}}
}

func (q WindowQuery) Type() QueryType {
	return QueryTypeWindow
}

func (q WindowQuery) Valid() bool {
	return q.By.Title.Valid()
}

type ColorQueryBy struct {
	Color RGBA
}

type ColorQuery struct {
	ID string
	By ColorQueryBy
}

func NewColorQuery(id string, color RGBA) ColorQuery {
	return ColorQuery{ID: id, By: ColorQueryBy{Color: color}}
}

func (q ColorQuery) Type() QueryType {
	return QueryTypeColor
}

func (q ColorQuery) Valid() bool {
	return true
}

type WindowElementQueryBy struct {
	Description WindowElementDescription
}

type WindowElementQuery struct {
	ID string
	By WindowElementQueryBy
}

func NewWindowElementQuery(id string, description WindowElementDescription) WindowElementQuery {
	return WindowElementQuery{ID: id, By: WindowElementQueryBy{Description: description}}
}

func (q WindowElementQuery) Type() QueryType {
	return QueryTypeWindowElement
}

func (q WindowElementQuery) Valid() bool {
	return q.By.Description.Valid()
}

func IsTextQuery(v any) bool {
	_, ok := v.(TextQuery)
	if ok {
		return true
	}
	_, ok = v.(*TextQuery)
	return ok
}

func IsWordQuery(v any) bool {
	switch q := v.(type) {
	case TextQuery:
		return q.IsWord()
	case *TextQuery:
		return q != nil && q.IsWord()
	default:
		return false
	}
}

func IsLineQuery(v any) bool {
	switch q := v.(type) {
	case TextQuery:
		return q.IsLine()
	case *TextQuery:
		return q != nil && q.IsLine()
	default:
		return false
	}
}

func IsWindowQuery(v any) bool {
	_, ok := v.(WindowQuery)
	if ok {
		return true
	}
	_, ok = v.(*WindowQuery)
	return ok
}

func IsColorQuery(v any) bool {
	_, ok := v.(ColorQuery)
	if ok {
		return true
	}
	_, ok = v.(*ColorQuery)
	return ok
}

func IsWindowElementQuery(v any) bool {
	_, ok := v.(WindowElementQuery)
	if ok {
		return true
	}
	_, ok = v.(*WindowElementQuery)
	return ok
}
