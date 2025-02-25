package types

// key - form's field, value - error msg
type errors map[string]string

func (e errors) Add(field, message string) {
	e[field] = message
}

func (e errors) Get(field string) string {
	em, ok := e[field]
	if !ok {
		return ""
	}
	return em
}
