package fauxy

// Proxy facilitates the connection(s) from one endpoint to anouther.
type Proxy interface {
	Start() error
	Stop()
}
