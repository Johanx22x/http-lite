package http

type Header map[string][]string

func (h Header) Set(key, value string) {
	h[key] = append(h[key], value)
}

func (h Header) Get(key string) string {
	if values, ok := h[key]; ok {
		return values[0]
	}
	return ""
}
