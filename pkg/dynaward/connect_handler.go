package dynaward

import (
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

type connectHandler struct {
	logger *slog.Logger
	fwd    *ForwardPool
}

func (ch *connectHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	t0 := time.Now()

	j, ok := w.(http.Hijacker)
	if !ok {
		ch.logger.Error("Error: writer not http.Hijacker", "path", r.URL.Path, "host", r.Host)
		http.Error(w, "Error: writer not http.Hijacker", http.StatusInternalServerError)
		return
	}

	conn, _, err := j.Hijack()
	if err != nil {
		ch.logger.Error("Error: "+err.Error(), "path", r.URL.Path, "host", r.Host)
		http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	fc, err := ch.fwd.ConnectionFor(ctx, r.Host)
	if err != nil {
		ch.logger.Error("Error: "+err.Error(), "path", r.URL.Path, "host", r.Host)
		http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ch.logger.Info("Connect", "host", r.Host, "pod_namespace", fc.Namespace, "pod_name", fc.PodName, "pod_port", fc.PodPort)

	id := string(uuid.NewUUID())
	hdr := http.Header{}
	hdr.Set(corev1.StreamType, corev1.StreamTypeError)
	hdr.Set(corev1.PortHeader, strconv.Itoa(fc.PodPort))
	hdr.Set(corev1.PortForwardRequestIDHeader, id)

	estream, err := fc.Conn.CreateStream(hdr)
	if err != nil {
		ch.logger.Error("cannot create error stream: "+err.Error(), "path", r.URL.Path, "host", r.Host)
		http.Error(w, "cannot create error stream: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_ = estream.Close()
	defer fc.Conn.RemoveStreams(estream)

	hdr.Set(corev1.StreamType, corev1.StreamTypeData)
	dstream, err := fc.Conn.CreateStream(hdr)
	if err != nil {
		ch.logger.Error("cannot create data stream: "+err.Error(), "path", r.URL.Path, "host", r.Host)
		http.Error(w, "cannot create data stream: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer fc.Conn.RemoveStreams(dstream)

	// proxy-level response
	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

	trace := NewRoundTripTrace()
	trace.Host = r.Host
	ch.fwd.tracestore.Add(id, trace)

	outbound := io.MultiWriter(dstream, trace.Request)
	inbound := io.MultiWriter(conn, trace.Response)

	// ch.logger.Info("Spawning bidirectional copier")
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		defer dstream.Close()
		_, err := io.Copy(outbound, conn)
		return err
	})
	eg.Go(func() error {
		_, err := io.Copy(inbound, dstream)
		return err
	})

	if err := eg.Wait(); err != nil {
		ch.logger.Error("error handling stream", "error", err)
	}
	ch.logger.Info("ok", "inbound_bytes", trace.Response.Len(), "outbound_bytes", trace.Request.Len(), "request_duration_seconds", time.Since(t0).Seconds())
	return
}
