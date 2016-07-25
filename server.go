package objets

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/dkumor/acmewrapper"
	"github.com/gorilla/mux"
	"github.com/tsileo/s3layer"
)

type Server struct {
	objets   *Objets
	shutdown chan struct{}
}

func NewServer(objets *Objets) *Server {
	return &Server{
		objets:   objets,
		shutdown: make(chan struct{}),
	}
}

func (s *Server) Shutdown() {
	s.shutdown <- struct{}{}
}

func (s *Server) Close() error {
	return s.objets.Close()
}

func (s *Server) Serve() error {
	go func() {
		m := mux.NewRouter()
		s4 := &s3layer.S4{s.objets}
		h := http.HandlerFunc(s4.Handler())
		m.Handle("/", h)
		m.Handle("/{bucket}", h)
		m.Handle("/{bucket}/", h)
		m.Handle("/{bucket}/{path:.+}", h)

		if s.objets.conf.AutoTLS {
			w, err := acmewrapper.New(acmewrapper.Config{
				Domains: s.objets.conf.Domains,
				Address: s.objets.conf.Listen(),

				TLSCertFile: filepath.Join(s.objets.conf.DataDir(), "cert.pem"),
				TLSKeyFile:  filepath.Join(s.objets.conf.DataDir(), "key.pem"),

				RegistrationFile: filepath.Join(s.objets.conf.DataDir(), "user.reg"),
				PrivateKeyFile:   filepath.Join(s.objets.conf.DataDir(), "user.pem"),

				TOSCallback: acmewrapper.TOSAgree,
			})

			if err != nil {
				panic(err)
			}

			tlsconfig := w.TLSConfig()

			listener, err := tls.Listen("tcp", s.objets.conf.Listen(), tlsconfig)
			if err != nil {
				panic(err)
			}

			// To enable http2, we need http.Server to have reference to tlsconfig
			// https://github.com/golang/go/issues/14374
			server := &http.Server{
				Addr:      s.objets.conf.Listen(),
				Handler:   m,
				TLSConfig: tlsconfig,
			}
			server.Serve(listener)
		} else {
			http.ListenAndServe(s.objets.conf.Listen(), m)
		}
	}()
	log.Printf("Listening on %s\n", s.objets.conf.Listen())
	s.tillShutdown()
	return s.Close()
}

func (s *Server) tillShutdown() {
	// Listen for shutdown signal
	cs := make(chan os.Signal, 1)
	signal.Notify(cs, os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	for {
		select {
		case <-cs:
			log.Printf("Shutting down...")
			return
		case <-s.shutdown:
			return
		}
	}
}
