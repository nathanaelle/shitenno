package frontend

// Nginx Handler
func Nginx() Handler {
	return &HttpHandler{}
}
