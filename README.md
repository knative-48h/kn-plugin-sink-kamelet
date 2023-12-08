# kn-plugin-sink-kamelet

# Requirements
The kn-plugin-sink-kamelet
- Knative Eventing and Serving
- Camel-k Operator

# Usage 
```shell
Plugin manages Kamelets and Pipes as Knative eventing sinks.

Usage:
  kn-sink-kamelet [command]

Available Commands:
  binding     Configure and manage a Kamelet binding.
  completion  Generate the autocompletion script for the specified shell
  describe    Show details of given Kamelet source type
  help        Help about any command
  list        List available Kamelet sink types
  version     Prints the plugin version

Flags:
  -h, --help   help for kn-sink-kamelet

Use "kn-sink-kamelet [command] --help" for more information about a command
```
# Commands

## binding
```shell
Configure and manage a Kamelet binding.

Usage:
  kn-sink-kamelet binding [command]

Examples:

  # Configure and manage a Kamelet binding.
  kn-sink-kamelet binding create|update|delete

Available Commands:
  create      Create Kamelet bindings and bind sink to Knative broker, channel or service.
  delete      Delete Kamelet binding by its name.

Flags:
  -h, --help   help for binding

Use "kn-sink-kamelet binding [command] --help" for more information about a command
```

### `binding create`
```shell
Usage:
  kn-sink-kamelet binding create NAME [flags]

Examples:

  # Create Kamelet binding with source and sink.
  kn source kamelet binding create NAME

  # Add a binding properties
  kn source kamelet binding create NAME --kamelet=name --sink|broker|channel|service=<name> --property=<key>=<value>

Flags:
      --broker string             Uses a broker as binding source.
      --broker-type string        Customize broker type provided to the binding source.
      --ce-override stringArray   Customize cloud events property in the form of "<key>=<value>"
      --ce-spec string            Customize cloud events spec version provided to the binding source.
      --ce-type string            Customize cloud events type provided to the binding source.
      --channel string            Uses a channel as binding source.
      --force                     Apply the changes even if the binding already exists.
  -h, --help                      help for create
      --kamelet string            Kamelet sink.
  -n, --namespace string          Specify the namespace to operate in.
      --property stringArray      Add a sink property in the form of "<key>=<value>"
      --service string            Uses a Knative service as binding source.
  -s, --source string             Source expression to define the binding sink.

```
### `binding delete`
```shell
Usage:
  kn-sink-kamelet binding delete NAME [flags]

Examples:

  # Delete Kamelet binding.
  kn source kamelet binding delete NAME

Flags:
  -h, --help               help for delete
  -n, --namespace string   Specify the namespace to operate in.
```

## describe
```shell
Usage:
  kn-sink-kamelet describe NAME [flags]

Examples:

  # Describe given Kamelets
  kn sink kamelet describe NAME

  # Describe given Kamelets in YAML output format
  kn sink kamelet describe NAME -o yaml

Flags:
      --allow-missing-template-keys   If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats. (default true)
  -h, --help                          help for describe
  -n, --namespace string              Specify the namespace to operate in.
  -o, --output string                 Output format. One of: json|yaml|name|go-template|go-template-file|template|templatefile|jsonpath|jsonpath-as-json|jsonpath-file|url.
      --show-managed-fields           If true, keep the managedFields when printing objects in JSON or YAML format.
      --template string               Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is golang templates [http://golang.org/pkg/text/template/#pkg-overview].
  -v, --verbose                       More output.

```
## list
```shell
Usage:
  kn-sink-kamelet list [flags]

Aliases:
  list, ls

Examples:

  # List available Kamelets
  kn sink kamelet list

  # List available Kamelets in YAML output format
  kn sink kamelet list -o yaml

Flags:
  -A, --all-namespaces                If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.
      --allow-missing-template-keys   If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats. (default true)
  -h, --help                          help for list
  -n, --namespace string              Specify the namespace to operate in.
      --no-headers                    When using the default output format, don't print headers (default: print headers).
  -o, --output string                 Output format. One of: (json, yaml, name, go-template, go-template-file, template, templatefile, jsonpath, jsonpath-as-json, jsonpath-file).
      --show-managed-fields           If true, keep the managedFields when printing objects in JSON or YAML format.
      --template string               Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is golang templates [http://golang.org/pkg/text/template/#pkg-overview].

```
