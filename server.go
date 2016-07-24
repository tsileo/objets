package objets

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/dkumor/acmewrapper"
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
	return nil
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

		// w, err := acmewrapper.New(acmewrapper.Config{
		// 	Domains: []string{"0.a4.io"},
		// 	Address: ":443",

		// 	TLSCertFile: "cert.pem",
		// 	TLSKeyFile:  "key.pem",

		// 	RegistrationFile: "user.reg",
		// 	PrivateKeyFile:   "user.pem",

		// 	TOSCallback: acmewrapper.TOSAgree,
		// })

		// if err != nil {
		// 	log.Fatal("acmewrapper: ", err)
		// }

		// tlsconfig := w.TLSConfig()

		// listener, err := tls.Listen("tcp", ":443", tlsconfig)
		// if err != nil {
		// 	log.Fatal("Listener: ", err)
		// }

		// // To enable http2, we need http.Server to have reference to tlsconfig
		// // https://github.com/golang/go/issues/14374
		// server := &http.Server{
		// 	Addr:      ":443",
		// 	Handler:   m,
		// 	TLSConfig: tlsconfig,
		// }
		// server.Serve(listener)

		http.ListenAndServe(":8060", m)
	}()
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
			return
		case <-s.shutdown:
			return
		}
	}
}
