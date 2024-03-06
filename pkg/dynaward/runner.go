package dynaward

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func (o *Options) Run(f cmdutil.Factory) error {
	sho := slog.HandlerOptions{}
	switch o.Verbosity {
	case InfoVerbosityLevel:
		sho.Level = slog.LevelInfo
	default:
		sho.Level = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(o.ErrOut, &sho))
	handler, closer, err := o.wrapServe(logger, f)
	if err != nil {
		return err
	}
	defer func() {
		logger.Info("Cleaning up")
		if closer != nil {
			closer()
		}
	}()

	logger.Info("Listening", "addr", o.Listen)
	return http.ListenAndServe(o.Listen, handler)
}

func (o *Options) controlIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	fwd := ExtractForwardPool(r)
	fwd.mut.RLock()
	defer fwd.mut.RUnlock()

	fmt.Fprintf(w, "[ %d active port-forwards ]\n", len(fwd.cache))
	for host, fc := range fwd.cache {
		fmt.Fprintf(w, "%s -> %s/%s:%d\n", host, fc.Namespace, fc.PodName, fc.PodPort)
	}
}

func (o *Options) wrapServe(logger *slog.Logger, f cmdutil.Factory) (http.HandlerFunc, func(), error) {
	kc, err := f.KubernetesClientSet()
	if err != nil {
		return nil, nil, err
	}

	rc, err := f.ToRESTConfig()
	if err != nil {
		return nil, nil, err
	}

	fwd := &ForwardPool{
		Client: kc,
		Config: rc,
		Logger: logger,
		cache:  map[string]*ForwardConnection{},
		mut:    sync.RWMutex{},
	}

	ctrl := http.NewServeMux()
	ctrl.HandleFunc("/", o.controlIndex)
	ctrl.HandleFunc("/favicon.ico", http.NotFound)

	hnd := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		t0 := time.Now()

		if r.Host == o.Listen {
			if o.Control {
				logger.Info("Received control request", "host", r.Host, "path", r.URL.Path)
				ctrl.ServeHTTP(w, InjectForwardPool(r, fwd))
			} else {
				logger.Error("Received control request, but control endpoint disabled", "host", r.Host)
				http.Error(w, "Control endpoint disabled", http.StatusForbidden)
			}
			return
		}

		logger.Info("Received proxy request", "path", r.URL.Path, "host", r.Host, "method", r.Method, "url", r.URL.String())

		if r.Method == http.MethodConnect {
			logger.Error("This proxy does not support 'CONNECT' yet", "path", r.URL.Path, "host", r.Host)
			http.Error(w, "This proxy does not support 'CONNECT' yet", http.StatusMethodNotAllowed)
			return
		}

		fc, err := fwd.ConnectionFor(ctx, r.Host)
		if err != nil {
			logger.Error("Error: "+err.Error(), "path", r.URL.Path, "host", r.Host)
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Info("Routing", "host", r.Host, "pod_namespace", fc.Namespace, "pod_name", fc.PodName, "pod_port", fc.PodPort)

		hdr := http.Header{}
		hdr.Set(corev1.StreamType, corev1.StreamTypeError)
		hdr.Set(corev1.PortHeader, strconv.Itoa(fc.PodPort))
		hdr.Set(corev1.PortForwardRequestIDHeader, string(uuid.NewUUID()))

		estream, err := fc.Conn.CreateStream(hdr)
		if err != nil {
			logger.Error("cannot create error stream: "+err.Error(), "path", r.URL.Path, "host", r.Host)
			http.Error(w, "cannot create error stream: "+err.Error(), http.StatusInternalServerError)
			return
		}
		_ = estream.Close()
		defer fc.Conn.RemoveStreams(estream)

		hdr.Set(corev1.StreamType, corev1.StreamTypeData)
		dstream, err := fc.Conn.CreateStream(hdr)
		if err != nil {
			logger.Error("cannot create data stream: "+err.Error(), "path", r.URL.Path, "host", r.Host)
			http.Error(w, "cannot create data stream: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer fc.Conn.RemoveStreams(dstream)

		r2 := r.Clone(ctx)
		_ = r2.Body.Close()

		if err := r2.Write(dstream); err != nil {
			logger.Error("cannot write request to data stream: "+err.Error(), "path", r.URL.Path, "host", r.Host)
			http.Error(w, "cannot write request to data stream: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := dstream.Close(); err != nil {
			logger.Error("cannot close data stream: "+err.Error(), "path", r.URL.Path, "host", r.Host)
			http.Error(w, "cannot close data stream: "+err.Error(), http.StatusInternalServerError)
			return
		}

		n, err := io.Copy(w, dstream)
		if err != nil {
			logger.Error("cannot copy response from data stream: "+err.Error(), "path", r.URL.Path, "host", r.Host)
			http.Error(w, "cannot copy response from data stream: "+err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Info("ok", "response_length_bytes", n, "request_duration_seconds", time.Since(t0).Seconds())
		return
	}

	return hnd, fwd.Close, nil
}
