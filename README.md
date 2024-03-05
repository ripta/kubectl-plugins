
kubectl 1.12 or newer is required. Place the plugin binaries in your `$PATH`,
and kubectl plugin commands will Just Work™.


## Installation

To install a specific plugin, e.g., `kubectl-dynaward`:

```
go install github.com/ripta/kubectl-plugins/cmd/kubectl-dynaward
```

To install **all plugins** as separate binaries:

```
go install github.com/ripta/kubectl-plugin/cmd/...
```


## `kubectl-dynaward`

Dynaward (dynamic port-forward) is an experimental HTTP proxy that allows you
to refer to `.cluster.local` addresses from outside of the cluster. It relies
on service names and performs its own mapping to pod names.

There is connection pooling, in that you will only incur lookup and connection
time for the first request to a hostname.

Routing is based on `Service` names and ports:

```
❯ kubectl -n echo get po
NAME                          READY   STATUS    RESTARTS   AGE
echoserver-5ccf76f46b-kmtmg   1/1     Running   0          3d4h
echoserver-5ccf76f46b-vqjtn   1/1     Running   0          3d4h

❯ kubectl -n echo get svc
NAME         TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
echoserver   ClusterIP   10.52.21.129   <none>        80/TCP    5y54d
```

Start up dynaward, which starts a proxy on `localhost:3128` by default:

```
❯ kubectl dynaward
{"time":"2024-03-04T01:20:50.628468-08:00","level":"INFO","msg":"Listening","addr":"localhost:3128"}
```

Then set your HTTP proxy (e.g., by setting `http_proxy` environment variable,
or by explicitly configuring your HTTP client). On curl, this is `-x`, while on
apachebench, it's `-X`:

```
❯ curl -x localhost:3128 echoserver.echo.svc.cluster.local:80

❯ ab -n 100 -c 5 -X localhost:3128 http://echoserver.echo.svc.cluster.local/
```

using `ab` on macOS, you may need to use `127.0.0.1` instead of `localhost`.

Limitations of dynaward:

1. does not yet detect disappearing pods
2. does not balance across multiple pods in a service
3. no control over internals
4. unaware of DNS, so it does not support pod A/AAAA records, pod hostnames, or
   pod subdomains
5. does not technically support custom cluster domains, but because dynaward
   looks at the first two segments of the hostname, it doesn't actually
   validate the cluster domain.
6. currently HTTP-only, and does not yet support CONNECT.

## Hyperbinary

All plugins are also available as a single hyperbinary that you can name
whatever you want. In the following example, it's installed as `$HYPERBINARY`:

```
HYPERBINARY=$HOME/bin/hypercmd
go install -o $HYPERBINARY github.com/ripta/kubectl-plugin/hyperbinary
```

Optionally, you can install symlinks to each command:

```
hypercmd --make-symlinks
```

Each symlink will be placed in the current working directory. Some help and
example pages may be wonky, e.g., examples will not be aware that it is in a
hyperbinary.


## Development

### Updating Dependencies

```
make update
make build
git add .
```

## Contributions

All contributions are welcome. Open a pull request, love 😘
