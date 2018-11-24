
This repository holds all my custom kubectl plugins. Usage depends on the
version of kubectl you have:

* If you have kubectl 1.12 or newer (recommended): place the plugin binary in
  your PATH, and kubectl plugin commands will just work.
* If you have kubectl 1.11 or older: see "Legacy Compatibility" section below.


## Installation

To install a specific plugin, e.g., `kubectl-ssh`:

```
go install github.com/ripta/kubectl-plugins/cmd/kubectl-ssh
```

To install **all plugins** as separate binaries:

```
go install github.com/ripta/kubectl-plugin/cmd/...
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

## Legacy Compatibility (1.11 or older)

If you're running kubectl <= 1.5, you should really upgrade.

If you're running kubectl 1.6 or 1.7, neither has support for plugins. In
theory, plugins are vendored with Kubernetes client-go v9.0, which supports
Kubernetes as far back as v1.6. Your mileage may vary, depending on featureset.

If you're running kubectl 1.8–1.11, you have two options:

- Craft a [plugin.yaml](https://v1-11.docs.kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/)
  for each plugin you would like to use. This is the official way in
  Kubernetes, but is tedious.
- If all your plugins are 1.12-compatible (which all plugins in this repository are),
  you can opt to emulate the call-style of 1.12 plugins by adding a function to your
  .bashrc like below, and then place the plugin binaries in your PATH.


```
function kubectl() {
  if [ "$#" == "0" ]; then
    command kubectl
  else
    cmd="$1"
    shift
    if command -v "kubectl-$cmd" 2>/dev/null 1>&2; then
      command kubectl-$cmd "$@"
    else
      command kubectl "$cmd" "$@"
    fi
  fi
}
```

## Contributions

All contributions are welcome. Open a pull request, love 😘
