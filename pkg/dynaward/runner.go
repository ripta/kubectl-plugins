package dynaward

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func (o *Options) Run(f cmdutil.Factory) error {
	//p := goproxy.NewProxyHttpServer()
	//p.Verbose = true
	//p.ConnectDialWithReq = func(req *http.Request, network string, addr string) (net.Conn, error) {
	//	fmt.Printf("Dial %s - %s\n", network, addr)
	//	return nil, nil
	//}

	//tr, _, err := spdy.RoundTripperFor()
	//p.Tr = tr

	logger := slog.New(slog.NewJSONHandler(o.ErrOut, &slog.HandlerOptions{
		AddSource: false,
	}))
	handler, err := o.wrapServe(logger, f)
	if err != nil {
		return err
	}

	addr := "localhost:3128"
	logger.Info("Listening", "addr", addr)
	return http.ListenAndServe(addr, handler)
}

func (o *Options) wrapServe(logger *slog.Logger, f cmdutil.Factory) (http.HandlerFunc, error) {
	kc, err := f.KubernetesClientSet()
	if err != nil {
		return nil, err
	}

	rc, err := f.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	hnd := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Info("Received request", "path", r.URL.Path, "host", r.Host, "method", r.Method, "url", r.URL.String())

		if r.Method == http.MethodConnect {
			http.Error(w, "This proxy does not support 'CONNECT' yet", http.StatusMethodNotAllowed)
			return
		}

		svcName, nsName, portStr, err := parseHost(r.Host)
		if err != nil {
			http.Error(w, "malformed host: "+err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Debug("Routing", "service_name", svcName, "namespace", nsName, "port", portStr)
		svc, err := kc.CoreV1().Services(nsName).Get(ctx, svcName, metav1.GetOptions{})
		if err != nil {
			http.Error(w, "cannot find service: "+err.Error(), http.StatusInternalServerError)
			return
		}

		portFound := 0
		portNum, _ := strconv.Atoi(portStr)
		for _, svcPort := range svc.Spec.Ports {
			if svcPort.Name == portStr || int(svcPort.Port) == portNum {
				portFound = svcPort.TargetPort.IntValue()
				break
			}
		}

		if portFound == 0 {
			http.Error(w, "cannot find port "+portStr+" on service "+r.Host, http.StatusInternalServerError)
			return
		}

		listOpts := metav1.ListOptions{
			LabelSelector: labels.SelectorFromSet(svc.Spec.Selector).String(),
		}
		pl, err := kc.CoreV1().Pods(nsName).List(ctx, listOpts)
		if err != nil {
			http.Error(w, "cannot list pods: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var targetPod *corev1.Pod
		for _, pod := range pl.Items {
			if pod.Status.Phase != corev1.PodRunning {
				continue
			}
			targetPod = &pod
			break
		}

		if targetPod == nil {
			http.Error(w, "did not find a healthy pod matching selector "+listOpts.LabelSelector, http.StatusInternalServerError)
			return
		}

		pfr := kc.RESTClient().Post().Prefix("api/v1").Namespace(targetPod.Namespace).Resource("pods").Name(targetPod.Name).SubResource("portforward")
		rt, up, err := spdy.RoundTripperFor(rc)
		if err != nil {
			http.Error(w, "cannot create roundtripper: "+err.Error(), http.StatusInternalServerError)
			return
		}

		hc := &http.Client{
			Transport: rt,
		}
		sd := spdy.NewDialer(up, hc, "POST", pfr.URL())
		conn, _, err := sd.Dial(portforward.PortForwardProtocolV1Name)
		if err != nil {
			logger.Error("Cannot dial", "error", err, "url", pfr.URL())
			http.Error(w, "cannot dial: "+err.Error(), http.StatusInternalServerError)
			return
		}

		defer conn.Close()

		hdr := http.Header{}
		hdr.Set(corev1.StreamType, corev1.StreamTypeError)
		hdr.Set(corev1.PortHeader, strconv.Itoa(portFound))
		hdr.Set(corev1.PortForwardRequestIDHeader, string(uuid.NewUUID()))

		estream, err := conn.CreateStream(hdr)
		if err != nil {
			http.Error(w, "cannot create error stream: "+err.Error(), http.StatusInternalServerError)
			return
		}
		estream.Close()
		defer conn.RemoveStreams(estream)

		hdr.Set(corev1.StreamType, corev1.StreamTypeData)
		dstream, err := conn.CreateStream(hdr)
		if err != nil {
			http.Error(w, "cannot create data stream: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.RemoveStreams(dstream)

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

	return hnd, nil
}

func parseHost(hostport string) (string, string, string, error) {
	host := hostport
	port := "80"
	if strings.Contains(host, ":") {
		h, p, err := net.SplitHostPort(hostport)
		if err != nil {
			return "", "", port, err
		}

		host = h
		port = p
	}

	if !strings.HasSuffix(host, ".cluster.local") {
		return "", "", port, fmt.Errorf("expected host %s to end in .cluster.local", host)
	}

	segs := strings.Split(host, ".")
	if len(segs) < 5 {
		return "", "", port, fmt.Errorf("expected at least 5 dot-separated segments, but found %d", len(segs))
	}

	return segs[0], segs[1], port, nil
}
