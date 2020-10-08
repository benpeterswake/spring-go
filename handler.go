package spring

type Handler struct {
	Route    string
	Handler  interface{}
	Params   []string
	Method   string
	Produces string
	Consumes string
}
