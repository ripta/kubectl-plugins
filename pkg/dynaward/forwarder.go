package dynaward

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type ForwardConnection struct {
	Conn httpstream.Connection

	Namespace   string
	ServiceName string
	ServicePort string
	PodName     string
	PodPort     int
}

type ForwardPool struct {
	Client *kubernetes.Clientset
	Config *rest.Config
	Logger *slog.Logger

	cache map[string]*ForwardConnection
	mut   sync.RWMutex
}

func (fwd *ForwardPool) Close() {
	fwd.mut.Lock()
	defer fwd.mut.Unlock()

	for hostport, fc := range fwd.cache {
		if err := fc.Conn.Close(); err != nil {
			fwd.Logger.Error("closing connection", "error", err, "host:port", hostport)
		}
	}
}

// ConnectionFor finds an existing connection to a pod, or creates it if one
// didn't already exist.
func (fwd *ForwardPool) ConnectionFor(ctx context.Context, hostport string) (*ForwardConnection, error) {
	// Fast path - read lock
	fwd.mut.RLock()
	conn, ok := fwd.cache[hostport]
	fwd.mut.RUnlock()
	if ok {
		fwd.Logger.Debug("Reusing connection", "host:port", hostport)
		return conn, nil
	}

	// Slow path - read/write lock
	fwd.mut.Lock()
	defer fwd.mut.Unlock()

	// Recheck cache in case another goroutine got here first
	conn, ok = fwd.cache[hostport]
	if ok {
		fwd.Logger.Debug("Reusing connection", "host:port", hostport)
		return conn, nil
	}

	fwd.Logger.Debug("Creating new connection", "host:port", hostport)
	conn, err := fwd.newConnectionFor(ctx, hostport)
	if err != nil {
		return nil, err
	}

	fwd.cache[hostport] = conn
	return conn, nil
}

// newConnectionFor creates a new forwarding connection for a specific hostport
func (fwd *ForwardPool) newConnectionFor(ctx context.Context, hostport string) (*ForwardConnection, error) {
	svcName, nsName, portStr, err := parseHost(hostport)
	if err != nil {
		return nil, fmt.Errorf("malformed host: %w", err)
	}

	fc := &ForwardConnection{
		Namespace:   nsName,
		ServiceName: svcName,
		ServicePort: portStr,
	}

	fwd.Logger.Debug("Routing", "service_name", svcName, "namespace", nsName, "port", portStr)
	svc, err := fwd.Client.CoreV1().Services(nsName).Get(ctx, svcName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot find service: %w", err)
	}

	portNum, _ := strconv.Atoi(portStr)
	for _, svcPort := range svc.Spec.Ports {
		if svcPort.Name == portStr || int(svcPort.Port) == portNum {
			fc.PodPort = svcPort.TargetPort.IntValue()
			break
		}
	}

	if fc.PodPort == 0 {
		return nil, errors.New("cannot find port " + portStr + " on service " + hostport)
	}

	listOpts := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(svc.Spec.Selector).String(),
	}
	pl, err := fwd.Client.CoreV1().Pods(nsName).List(ctx, listOpts)
	if err != nil {
		return nil, fmt.Errorf("cannot list pods: %w", err)
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
		return nil, fmt.Errorf("did not find a healthy pod matching selector %q", listOpts.LabelSelector)
	}
	fc.PodName = targetPod.Name

	pfr := fwd.Client.RESTClient().Post().Prefix("api/v1").Namespace(targetPod.Namespace).Resource("pods").Name(targetPod.Name).SubResource("portforward")
	rt, up, err := spdy.RoundTripperFor(fwd.Config)
	if err != nil {
		return nil, fmt.Errorf("cannot create roundtripper: %w", err)
	}

	hc := &http.Client{
		Transport: rt,
	}
	sd := spdy.NewDialer(up, hc, "POST", pfr.URL())
	conn, _, err := sd.Dial(portforward.PortForwardProtocolV1Name)
	if err != nil {
		return nil, fmt.Errorf("cannot dial %s: %w", pfr.URL().String(), err)
	}

	fc.Conn = conn
	return fc, nil
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
