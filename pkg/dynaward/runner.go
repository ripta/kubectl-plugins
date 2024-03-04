package dynaward

import (
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func (o *Options) Run(f cmdutil.Factory) error {
	logger := slog.New(slog.NewJSONHandler(o.ErrOut, &slog.HandlerOptions{
		AddSource: false,
	}))
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

	addr := "localhost:3128"
	logger.Info("Listening", "addr", addr)
	return http.ListenAndServe(addr, handler)
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

	hnd := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Info("Received request", "path", r.URL.Path, "host", r.Host, "method", r.Method, "url", r.URL.String())

		if r.Method == http.MethodConnect {
			http.Error(w, "This proxy does not support 'CONNECT' yet", http.StatusMethodNotAllowed)
			return
		}

		fc, err := fwd.ConnectionFor(ctx, r.Host)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		hdr := http.Header{}
		hdr.Set(corev1.StreamType, corev1.StreamTypeError)
		hdr.Set(corev1.PortHeader, strconv.Itoa(fc.PodPort))
		hdr.Set(corev1.PortForwardRequestIDHeader, string(uuid.NewUUID()))

		estream, err := fc.Conn.CreateStream(hdr)
		if err != nil {
			http.Error(w, "cannot create error stream: "+err.Error(), http.StatusInternalServerError)
			return
		}
		_ = estream.Close()
		defer fc.Conn.RemoveStreams(estream)

		hdr.Set(corev1.StreamType, corev1.StreamTypeData)
		dstream, err := fc.Conn.CreateStream(hdr)
		if err != nil {
			http.Error(w, "cannot create data stream: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer fc.Conn.RemoveStreams(dstream)

		r2 := r.Clone(ctx)
		r2.Body.Close()

		if err := r2.Write(dstream); err != nil {
			http.Error(w, "cannot write request to data stream: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := dstream.Close(); err != nil {
			http.Error(w, "cannot close data stream: "+err.Error(), http.StatusInternalServerError)
			return
		}

		n, err := io.Copy(w, dstream)
		if err != nil {
			http.Error(w, "cannot copy response from data stream: "+err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Info("ok", "response_length_bytes", n)
		return
	}

	return hnd, fwd.Close, nil
}
