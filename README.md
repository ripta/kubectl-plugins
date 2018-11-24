
This repository holds all my custom kubectl plugins. Usage depends on the
version of kubectl you have:

* If you have kubectl 1.12 or newer (recommended): place the plugin binary in
  your PATH, and kubectl plugin commands will just work.
* If you have kubectl 1.11 or older: add a function to your .bashrc to override
  the kubectl command.


## Installation

To install a specific plugin, e.g., `kubectl-ssh`:

```
go install github.com/ripta/kubectl-plugins/cmd/kubectl-ssh
```

To install all plugins as separate binaries:

```
go install github.com/ripta/kubectl-plugin/cmd/...
```

