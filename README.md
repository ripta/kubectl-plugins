
kubectl 1.12 or newer is required. Place the plugin binaries in your `$PATH`,
and kubectl plugin commands will Just Work™.


## Installation

To install a specific plugin, e.g., `kubectl-ssh`:

```
go install github.com/ripta/kubectl-plugins/cmd/kubectl-ssh
```

To install **all plugins** as separate binaries:

```
go install github.com/ripta/kubectl-plugin/cmd/...
```

To use the examples included in this repo, add this to the preferences section
of your Dockerfile:

```
preferences:
  extensions:
    - name: ShowConfig
      extension:
        apiVersion: k.r8y.dev/v1alpha1
        kind: ShowConfig
        searchPaths:
          - $PATH_TO_REPO/examples
```

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
