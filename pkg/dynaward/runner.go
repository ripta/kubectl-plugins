package dynaward

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
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

	fmt.Fprint(w, "<!DOCTYPE html>\n")
	fmt.Fprint(w, "<html><body>\n")

	fmt.Fprintf(w, "<strong>%d active port-forwards</strong><br>\n", len(fwd.cache))
	fmt.Fprint(w, "<ul>\n")
	for host, fc := range fwd.cache {
		fmt.Fprintf(w, "<li>%s -> %s/%s:%d</li>\n", host, fc.Namespace, fc.PodName, fc.PodPort)
	}
	fmt.Fprint(w, "</ul>\n")

	traces := fwd.tracestore.List()
	sort.Strings(traces)
	fmt.Fprintf(w, "<strong>%d traces (%s)</strong><br>\n", len(traces), `<a href="/traces/clear">clear</a>`)
	fmt.Fprint(w, "<ul>\n")
	for _, id := range traces {
		trace := fwd.tracestore.Get(id)
		fmt.Fprintf(w, `<li><a href="/traces/%s">%s</a>: %d bytes request, %d bytes response</li>`, id, id, trace.Request.Len(), trace.Response.Len())
	}
	fmt.Fprint(w, "</ul>\n")
}

func (o *Options) controlClearTrace(w http.ResponseWriter, r *http.Request) {
	fwd := ExtractForwardPool(r)
	fwd.mut.RLock()
	defer fwd.mut.RUnlock()

	fwd.tracestore.Clear()
	http.Redirect(w, r, "/", http.StatusFound)
}

func (o *Options) controlViewTrace(w http.ResponseWriter, r *http.Request) {
	id, ok := strings.CutPrefix(r.URL.Path, "/traces/")
	if !ok {
		http.NotFound(w, r)
		return
	}

	fwd := ExtractForwardPool(r)
	fwd.mut.RLock()
	defer fwd.mut.RUnlock()

	trace := fwd.tracestore.Get(id)
	if trace == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(200)

	fmt.Fprintf(w, "Request:\n%s\n\n", trace.Request)
	fmt.Fprintf(w, "---\n\n")
	fmt.Fprintf(w, "Response:\n%s\n\n", trace.Response)
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
		Client:     kc,
		Config:     rc,
		Logger:     logger,
		cache:      map[string]*ForwardConnection{},
		mut:        sync.RWMutex{},
		tracestore: NewTraceStore(20),
	}

	ctrl := http.NewServeMux()
	ctrl.HandleFunc("/", o.controlIndex)
	ctrl.HandleFunc("/favicon.ico", http.NotFound)
	ctrl.HandleFunc("/traces/", o.controlViewTrace)
	ctrl.HandleFunc("/traces/clear", o.controlClearTrace)

	ch := connectHandler{
		logger: logger,
		fwd:    fwd,
	}

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
			ch.Handle(w, r)
			return
		}

		fc, err := fwd.ConnectionFor(ctx, r.Host)
		if err != nil {
			logger.Error("Error: "+err.Error(), "path", r.URL.Path, "host", r.Host)
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Info("Routing", "host", r.Host, "pod_namespace", fc.Namespace, "pod_name", fc.PodName, "pod_port", fc.PodPort)

		id := string(uuid.NewUUID())
		hdr := http.Header{}
		hdr.Set(corev1.StreamType, corev1.StreamTypeError)
		hdr.Set(corev1.PortHeader, strconv.Itoa(fc.PodPort))
		hdr.Set(corev1.PortForwardRequestIDHeader, id)

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

		trace := NewRoundTripTrace()
		trace.Host = r.Host
		fwd.tracestore.Add(id, trace)

		outbound := io.MultiWriter(dstream, trace.Request)
		inbound := io.MultiWriter(w, trace.Response)

		if err := r2.Write(outbound); err != nil {
			logger.Error("cannot write request to data stream: "+err.Error(), "path", r.URL.Path, "host", r.Host)
			http.Error(w, "cannot write request to data stream: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := dstream.Close(); err != nil {
			logger.Error("cannot close data stream: "+err.Error(), "path", r.URL.Path, "host", r.Host)
			http.Error(w, "cannot close data stream: "+err.Error(), http.StatusInternalServerError)
			return
		}

		n, err := io.Copy(inbound, dstream)
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
