---
title: "Debugging With Skaffold"
linkTitle: "Debugging"
weight: 30
featureId: debug
aliases: [/docs/how-tos/debug]
---

`skaffold debug` acts like `skaffold dev`, but it configures containers in pods
for debugging as required for each container's runtime technology.
The associated debugging ports are exposed and labelled so that they can be port-forwarded to the
local machine.  IDEs like [Google's Cloud Code extensions](https://cloud.google.com/code) use Skaffold's events
to automatically configure debug sessions.

One notable difference from `skaffold dev` is that `debug` disables image rebuilding and
syncing as it leads to users accidentally terminating debugging sessions by saving file changes.
These behaviours can be re-enabled with the `--auto-build`, `--auto-deploy`, and `--auto-sync`
flags.

## How It works

`skaffold debug` examines the built artifacts to determine the underlying language runtime technology.
Kubernetes manifests that reference these artifacts are transformed on-the-fly to enable the
language runtime's debugging functionality.  These transforms add or alter environment variables
and entrypoints, and more.

Some language runtimes require additional support files to enable debugging.
For these languages, a special set of [runtime-specific images](https://github.com/GoogleContainerTools/container-debug-support)
are configured as _init-containers_ to populate a shared-volume that is mounted into
each of the appropriate containers.  These images are hosted at
`gcr.io/k8s-skaffold/skaffold-debug-support`; alternative locations can be
specified in [Skaffold's global configuration]({{< relref "/docs/design/global-config.md" >}}).

`debug` makes some other adjustments to simplify the debug experience:

  - *Replica Counts*: `debug` rewrites  the replica counts to 1 for
    deployments, replica sets, and stateful sets.  This results in
    requests being considered one at a time.

  - *Kubernetes Probes*:  `debug` changes the timeouts on HTTP-based
    [liveness, readiness, and startup probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
    to 600 seconds (10 minutes) from the default of 1 second. 
    This change allows probes to be debugged, and avoids negative
    consequences from blocked probes when the app is already suspended
    during a debugging session.
    Failed liveness probes in particular result in the container
    being terminated and restarted.

	The probe timeout value can be set on a per-podspec basis by setting
	a `debug.cloud.google.com/probe/timeouts` annotation on the podspec's metadata
	with a valid duration (see [Go's time.ParseDuration()](https://pkg.go.dev/time#ParseDuration)).
    This probe timeout-rewriting can be skipped entirely by using `skip`.  For example:
    ```yaml
    metadata:
      annotations:
        debug.cloud.google.com/probe/timeouts: skip
    spec: ...
    ```

### Supported Language Runtimes

Debugging is currently supported for:
  - Go (runtime ID: `go`)
  - NodeJS (runtime ID: `nodejs`)
  - Java and JVM languages (runtime ID: `jvm`)
  - Python (runtime ID: `python`)
  - .NET Core (runtime ID: `netcore`)
  
Note that many debuggers may require additional information for the location of source files.
We are looking for ways to identify this information and to pass it back if found.

#### Go

Go-based applications are configured to run under [Delve](https://github.com/go-delve/delve) in its headless-server mode.
In order to configure your application for debugging, your app must be:

  - Identified as being Go-based by setting one of the [standard Go runtime
    environment variables](https://godoc.org/runtime) in the container, such as `GODEBUG`, `GOGC`, `GOMAXPROCS`,
    or `GOTRACEBACK`. `GOTRACEBACK=single` is the default setting for Go, and `GOTRACEBACK=all` is a 
    generally useful configuration.
  - Built with the `-gcflags='all=-N -l'` options to disable optimizations and inlining.
    Debugging can be confusing otherwise due to seemingly-random
    execution jumps from statement reordering and inlining.
    Skaffold [_Profiles_]({{< relref "/docs/environment/profiles.md" >}}) are a useful option.
    Note that the `golang:NN-alpine` container images do not include a C compiler which is required
    for `-gcflags='all=-N -l'`.

Note for users of [VS Code's debug adapter for Go](https://github.com/Microsoft/vscode-go): the debug adapter
may require configuring both the _local_ and _remote_ source path prefixes via the `cwd` and `remotePath` properties.
The `cwd` property should point to the top-level container of your source files and should generally match
the artifact's `context` directory in the `skaffold.yaml`.  The `remotePath` path property should be set to the
remote source location _during compilation_.  For example, the `golang` images, which are
[often used in multi-stage builds](https://github.com/GoogleContainerTools/skaffold/tree/master/examples/getting-started/Dockerfile),
copy the source code to `/go`.  The following
[remote launch configuration](https://github.com/Microsoft/vscode-go/wiki/Debugging-Go-code-using-VS-Code#remote-debugging)
works in this case:
```json
{
  "name": "Skaffold Debug",
  "type": "go",
  "request": "launch",
  "mode": "remote",
  "host": "localhost",
  "port": 56268,
  "cwd": "${workspaceFolder}",
  "remotePath": "/go/"
}
```

#### Java and Other JVM Languages

Java/JVM applications are configured to expose the JDWP agent using the `JAVA_TOOL_OPTIONS`
environment variable.  
Note that the use of `JAVA_TOOL_OPTIONS` causes extra debugging output from the JVM on launch.

#### NodeJS

NodeJS applications are configured to use the Chrome DevTools inspector via the `--inspect` argument.

Note that the client must first obtain [the inspector UUID](https://github.com/nodejs/node/issues/9185#issuecomment-254872466).

{{< alert title="Note" >}}
Many applications use NodeJS-based tools as part of their launch, like <tt>npm</tt>, rather than
invoke <tt>node</tt> directly.  These intermediate <tt>node</tt> instances may interpret the
<tt>--inspect</tt> arguments.  Skaffold introduces a <tt>node</tt> wrapper that
only invokes the real <tt>node</tt> with <tt>--inspect</tt> if running an application script,
and skips scripts located in <tt>node_modules</tt>.  For more details see the
<a href="https://github.com/GoogleContainerTools/container-debug-support/pull/34">associated PR</a>.
{{< /alert >}}
  
#### Python

Python applications are configured to use [`ptvsd`](https://github.com/microsoft/ptvsd/), a
wrapper around [`pydevd`](https://github.com/fabioz/PyDev.Debugger) that uses the
[_debug adapter protocol_ (DAP)](https://microsoft.github.io/debug-adapter-protocol/). 

The DAP is supported by Visual Studio Code, [Eclipse LSP4e](https://projects.eclipse.org/projects/technology.lsp4e),
[and other editors and IDEs](https://microsoft.github.io/debug-adapter-protocol/implementors/tools/).
DAP is not yet supported by JetBrains IDEs like PyCharm.

#### .NET Core

.NET Core applications are configured to be deployed along with `vsdbg`.

In order to configure your application for debugging, your app must be:

- Identified as being dotnet-based by having an entrypoint using [dotnet](https://github.com/dotnet/sdk) cli
  or one of the following environment variables `ASPNETCORE_URLS`, `DOTNET_RUNNING_IN_CONTAINER`,
  `DOTNET_SYSTEM_GLOBALIZATION_INVARIANT`.
- Built with the `--configuration Debug` options to disable optimizations.

**Note for users of [VS Code's debug adapter for C#](https://github.com/OmniSharp/omnisharp-vscode):**
the following configuration can be used to debug a container. It assumes that your code is deployed
in `/app` or `/src` folder in the container. If that is not the case, the `sourceFileMap` property
should be changed to match the correct folder. `processId` is usually 1 but might be different if you
have an unusual entrypoint. You can also use `"${command:pickRemoteProcess}"` instead if supported by
your base image.  (`//` comments must be stripped.)
```json
{
    "name": "Skaffold Debug",
    "type": "coreclr",
    "request": "attach",
    "processId" : 1, 
    "justMyCode": true, // set to `true` in debug configuration and `false` in release configuration
    "pipeTransport": {
        "pipeProgram": "kubectl",
        "pipeArgs": [
            "exec",
            "-i",
            "<NAME OF YOUR POD>", // name of the pod you debug.
            "--"
        ],
        "pipeCwd": "${workspaceFolder}",
        "debuggerPath": "/dbg/netcore/vsdbg", // location where vsdbg binary installed.
        "quoteArgs": false
    },
    "sourceFileMap": {
        // Change this mapping if your app in not deployed in /src or /app in your docker image
        "/src": "${workspaceFolder}",
        "/app": "${workspaceFolder}"
        // May also be like this, depending of your repository layout
        // "/src": "${workspaceFolder}/src",
        // "/app": "${workspaceFolder}/src/<YOUR PROJECT TO DEBUG>"
    }
}
```

## IDE Support via Events and Metadata

`debug` provides additional support for IDEs to detect the debuggable containers and to determine
appropriate configuration parameters.

### Workload Annotations

Each transformed workload object carries a `debug.cloud.google.com/config` annotation with
a JSON object describing the debug configurations for the pod's containers (linebreaks for readability):
```  
	debug.cloud.google.com/config={
		"<containerName>":{"runtime":"<runtimeId>",...},
		"<containerName>":{"runtime":"<runtimeId>",...},
		}
```

For example the following annotation indicates that the container named `web` is a Go application
that is being debugged by a headless Delve session on port `56268` (linebreaks for readability):
```
debug.cloud.google.com/config={
  "web":{
    "artifact":"gcr.io/random/image",
    "runtime":"go",
    "ports":{"dlv":56268},
    "workingDir":"/some/path"}}
```

`artifact` is the corresponding artifact's image name in the `skaffold.yaml`.
`runtime` is the language runtime detected (one of: `go`, `jvm`, `nodejs`, `python`).
`ports` is a list of debug ports keyed by the language runtime debugging protocol.
`workingDir` is the working directory (if not an empty string).


### API: Events

Each debuggable container being started or stopped raises a _debug-container-event_ through
Skaffold's event mechanism ([gRPC](../references/api/grpc/#debuggingcontainerevent), 
[REST](../references/api/swagger/#/SkaffoldService/Events)).

<details>
<summary>`/v1/events` stream of `skaffold debug` within `examples/jib`</summary>

In this example, we do a `skaffold debug`, and then kill the deployed pod.  The deployment starts a new pod.  We get a terminated event for the container for the killed pod.

```json
{
  "result": {
    "timestamp": "2020-02-05T03:27:30.114354Z",
    "event": {
      "debuggingContainerEvent": {
        "status": "Started",
        "podName": "web-f6d56bcc5-6csgs",
        "containerName": "web",
        "namespace": "default",
        "artifact": "skaffold-jib",
        "runtime": "jvm",
        "debugPorts": {
          "jdwp": 5005
        }
      }
    },
    "entry": "Debuggable container started pod/web-f6d56bcc5-6csgs:web (default)"
  }
}
```

</details>



### API: State

The API's _state_ ([gRPC](../references/api/grpc/#skaffoldservice), [REST](../references/api/swagger/#/SkaffoldService/GetState)) also includes a list of debuggable containers.

<details>
<summary>The `/v1/state` listing debugging containers</summary>

```json
{
  "buildState": {
    "artifacts": {
      "skaffold-jib": "Complete"
    }
  },
  "deployState": {
    "status": "Complete"
  },
  "forwardedPorts": {
    "5005": {
      "localPort": 5005,
      "remotePort": 5005,
      "podName": "web-f6d56bcc5-6csgs",
      "containerName": "web",
      "namespace": "default",
      "portName": "jdwp",
      "resourceType": "pod",
      "resourceName": "web-f6d56bcc5-6csgs",
      "address": "127.0.0.1"
    },
    "8080": {
      "localPort": 8080,
      "remotePort": 8080,
      "namespace": "default",
      "resourceType": "service",
      "resourceName": "web",
      "address": "127.0.0.1"
    },
    "8081": {
      "localPort": 8081,
      "remotePort": 8080,
      "podName": "web-f6d56bcc5-6csgs",
      "containerName": "web",
      "namespace": "default",
      "resourceType": "pod",
      "resourceName": "web-f6d56bcc5-6csgs",
      "address": "127.0.0.1"
    }
  },
  "statusCheckState": {
    "status": "Not Started"
  },
  "fileSyncState": {
    "status": "Not Started"
  },
  "debuggingContainers": [
    {
      "status": "Started",
      "podName": "web-f6d56bcc5-6csgs",
      "containerName": "web",
      "namespace": "default",
      "artifact": "skaffold-jib",
      "runtime": "jvm",
      "debugPorts": {
        "jdwp": 5005
      }
    }
  ]
}

```

</details>


## Limitations

`skaffold debug` has some limitations.

### Unsupported Container Entrypoints

`skaffold debug` requires being able to examine and alter the
command-line used in the container entrypoint.  This transformation
will not work with images that use intermediate launch scripts or
binaries.

### Supported Deployers

`skaffold debug` is only supported with the `kubectl`, `kustomize`, and `helm` deployers.

{{< alert title="Note" >}}
Helm support requires using Helm v3.1.0 or greater.
{{< /alert >}}


### Deprecated Workload API Objects

`skaffold debug` does not support deprecated versions of Workload API objects:

  - `extensions/v1beta1` and `apps/v1beta1` was [deprecated in Kubernetes 1.8](https://github.com/kubernetes/kubernetes/blob/HEAD/CHANGELOG/CHANGELOG-1.8.md#other-notable-changes-16)
    and [removed in Kubernetes 1.16](https://kubernetes.io/blog/2019/07/18/api-deprecations-in-1-16/).
  - `apps/v1beta2` was [deprecated in Kubernetes 1.9](https://github.com/kubernetes/kubernetes/blob/HEAD/CHANGELOG/CHANGELOG-1.9.md#apps)
    and [removed in Kubernetes 1.16](https://kubernetes.io/blog/2019/07/18/api-deprecations-in-1-16/).

Applications should transition to the `apps/v1` APIs,
[introduced in Kubernetes 1.9](https://kubernetes.io/blog/2017/12/kubernetes-19-workloads-expanded-ecosystem/#workloads-api-ga).
