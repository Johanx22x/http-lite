package http

type Header map[string][]string

// Set establece el valor de un encabezado.
func (h Header) Set(key, value string) {
	h[key] = []string{value}
}

// Get obtiene el primer valor de un encabezado.
func (h Header) Get(key string) string {
	if values, ok := h[key]; ok && len(values) > 0 {
		return values[0]
	}
	return ""
}
