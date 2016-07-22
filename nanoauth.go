// Package nanoauth provides a uniform means of serving HTTP/S for golang
// projects securely. It allows the specification of a certificate (or
// generates one) as well as an auth token which is checked before the request
// is processed.
package nanoauth

import (
	"crypto/tls"
	"net"
	"net/http"
)

// Auth is a structure containing listener information
type Auth struct {
	child         http.Handler     // child is the http handler passed in
	Header        string           // Header is the authentication token's header name
	Certificate   *tls.Certificate // Certificate is the tls.Certificate to serve requests with
	ExcludedPaths []string         // ExcludedPaths is a list of paths to be excluded from being authenticated
	Token         string           // Token is the security/authentication string to validate by
}

var (
	// DefaultAuth is the default Auth object
	DefaultAuth = &Auth{}

	// whether or not to check auth tokens
	check = true
)

func init() {
	DefaultAuth.Header = "X-NANOBOX-TOKEN"
	DefaultAuth.Certificate, _ = Generate("nanobox.io")
}

// ServeHTTP is to implement the http.Handler interface. Also let clients know
// when I have no matching route listeners
func (self Auth) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	reqPath := req.URL.Path

	for _, path := range self.ExcludedPaths {
		if path == reqPath {
			check = false
			break
		}
	}

	if check {
		auth := ""
		if auth = req.Header.Get(self.Header); auth == "" {
			// check form value (case sensitive) if header not set
			auth = req.FormValue(self.Header)
		}

		if auth != self.Token {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	self.child.ServeHTTP(rw, req)
}

// ListenAndServeTLS starts a TLS listener and handles serving https
func (self *Auth) ListenAndServeTLS(addr, token string, h http.Handler, excludedPaths ...string) error {
	config := &tls.Config{
		Certificates: []tls.Certificate{*self.Certificate},
	}
	config.BuildNameToCertificate()
	tlsListener, err := tls.Listen("tcp", addr, config)
	if err != nil {
		return err
	}

	if token == "" {
		check = false
	}
	self.ExcludedPaths = excludedPaths
	self.Token = token

	if h == nil {
		h = http.DefaultServeMux
	}
	self.child = h

	return http.Serve(tlsListener, self)
}

// ListenAndServe starts a normal tcp listener and handles serving http while
// still validating the auth token.
func (self *Auth) ListenAndServe(addr, token string, h http.Handler, excludedPaths ...string) error {
	httpListener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	if token == "" {
		check = false
	}
	self.ExcludedPaths = excludedPaths
	self.Token = token

	if h == nil {
		h = http.DefaultServeMux
	}
	self.child = h

	return http.Serve(httpListener, self)
}

// ListenAndServeTLS is a shortcut function which uses the default one
func ListenAndServeTLS(addr, token string, h http.Handler, excludedPaths ...string) error {
	return DefaultAuth.ListenAndServeTLS(addr, token, h, excludedPaths...)
}

// ListenAndServe is a shortcut function which uses the default one
func ListenAndServe(addr, token string, h http.Handler, excludedPaths ...string) error {
	return DefaultAuth.ListenAndServe(addr, token, h, excludedPaths...)
}
