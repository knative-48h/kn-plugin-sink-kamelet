/*
 * Copyright © 2021 The Knative Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	camelv1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	camelkv1 "github.com/apache/camel-k/v2/pkg/client/camel/clientset/versioned/typed/camel/v1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	knerrors "knative.dev/client-pkg/pkg/errors"

	"path/filepath"
	"strings"

	"knative.dev/client-pkg/pkg/kn/commands"
)

var bindingCreateExample = `
  # Create Kamelet binding with source and sink.
  kn source kamelet binding create NAME

  # Add a binding properties
  kn source kamelet binding create NAME --kamelet=name --sink|broker|channel|service=<name> --property=<key>=<value>`

// newBindingCommand implements 'kn-sink-kamelet binding' command
func newBindingCreateCommand(p *KameletPluginParams) *cobra.Command {

	var properties []string
	var source string
	var sink string
	var broker string
	var channel string
	var service string
	//var kameletNamespace string
	var cloudEventsOverride []string
	var cloudEventsSpecVersion string
	var cloudEventsType string
	var sourceBrokerType string
	var force bool

	cmd := &cobra.Command{
		Use:     "create NAME",
		Short:   "Create Kamelet bindings and bind sink to Knative broker, channel or service.",
		Example: bindingCreateExample,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'kn-sink-kamelet binding create' requires the binding name as argument")
			}
			name := args[0]

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			client, err := p.NewKameletClient()
			if err != nil {
				return err
			}

			options := CreateBindingOptions{
				Name:                   name,
				SinkProperties:         properties,
				Sink:                   sink,
				Source:                 source,
				SourceBrokerType:       sourceBrokerType,
				CloudEventsOverride:    cloudEventsOverride,
				CloudEventsSpecVersion: cloudEventsSpecVersion,
				CloudEventsType:        cloudEventsType,
				Broker:                 broker,
				Channel:                channel,
				Service:                service,
				Force:                  force,
			}
			err = createBinding(client, p.Context, namespace, options)
			if err != nil {
				return err
			}

			return nil
		},
	}
	flags := cmd.Flags()
	commands.AddNamespaceFlags(flags, false)

	flags.StringVar(&sink, "kamelet", "", "Kamelet sink.")
	flags.StringVarP(&source, "source", "s", "", "Source expression to define the binding sink.")
	flags.StringVar(&broker, "broker", "", "Uses a broker as binding source.")
	flags.StringVar(&channel, "channel", "", "Uses a channel as binding source.")
	flags.StringVar(&service, "service", "", "Uses a Knative service as binding source.")
	flags.BoolVar(&force, "force", false, "Apply the changes even if the binding already exists.")
	flags.StringArrayVar(&properties, "property", nil, `Add a sink property in the form of "<key>=<value>"`)
	flags.StringVar(&cloudEventsSpecVersion, "ce-spec", "", "Customize cloud events spec version provided to the binding source.")
	flags.StringVar(&cloudEventsType, "ce-type", "", "Customize cloud events type provided to the binding source.")
	flags.StringVar(&sourceBrokerType, "broker-type", "", "Customize broker type provided to the binding source.")
	//flags.StringVarP(&kameletNamespace, "kamelet-namespace", "kn", "", "Namespace where the kamelet is located.")
	flags.StringArrayVar(&cloudEventsOverride, "ce-override", nil, `Customize cloud events property in the form of "<key>=<value>"`)
	return cmd

}

func createBinding(client camelkv1.CamelV1Interface, ctx context.Context, namespace string, options CreateBindingOptions) error {

	kamelet, err := client.Kamelets(namespace).Get(ctx, options.Name, v1.GetOptions{})
	if err != nil {
		knerrors.GetError(err)
	}

	sinkProps, err := parseProperties(options.SinkProperties)
	if err != nil {
		return knerrors.GetError(err)
	}
	sinkEnpointProps, err := asEndpointProperties(sinkProps)
	if err != nil {
		return knerrors.GetError(err)
	}
	sinkEndpoint := camelv1.Endpoint{
		Ref: &corev1.ObjectReference{
			Kind:       camelv1.KameletKind,
			APIVersion: camelv1.SchemeGroupVersion.String(),
			Name:       kamelet.Name,
			Namespace:  kamelet.Namespace,
		},
		Properties: &sinkEnpointProps,
	}
	if err := verifyProperties(kamelet, sinkEndpoint); err != nil {
		return knerrors.GetError(err)
	}

	var sourceRef corev1.ObjectReference
	if options.Source != "" {
		sourceRef, err = decodeSource(options.Source)
	} else if options.Broker != "" {
		sourceRef, err = decodeSource("broker:" + options.Broker)
	} else if options.Channel != "" {
		sourceRef, err = decodeSource("channel:" + options.Channel)
	} else if options.Service != "" {
		sourceRef, err = decodeSource("ksvc:" + options.Service)
	} else {
		err = fmt.Errorf("missing sink for binding - please use one of --sink, --broker, --channel, --service")
	}

	if err != nil {
		return knerrors.GetError(err)
	}

	if sourceRef.Namespace == "" {
		sourceRef.Namespace = namespace
	}

	sourceProps, err := getSourceProperties(options)
	if err != nil {
		return knerrors.GetError(err)
	}
	sinkEndpointProps, err := asEndpointProperties(sourceProps)
	if err != nil {
		return knerrors.GetError(err)
	}
	sourceEndpoint := camelv1.Endpoint{
		Properties: &sinkEndpointProps,
		Ref:        &sourceRef,
	}
	name := nameFor(options.Name, options.Sink, sourceRef)

	binding := camelv1.Pipe{
		ObjectMeta: v1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: camelv1.PipeSpec{
			Source: sourceEndpoint,
			Sink:   sinkEndpoint,
		},
	}

	existed := false
	_, err = client.Pipes(namespace).Create(ctx, &binding, v1.CreateOptions{})
	if err != nil && k8serrors.IsAlreadyExists(err) {
		if options.Force {
			existed = true

			existing, err := client.Pipes(namespace).Get(ctx, binding.Name, v1.GetOptions{})
			if err != nil {
				return knerrors.GetError(err)
			}
			// Update the custom resource
			binding.ResourceVersion = existing.ResourceVersion
			_, err = client.Pipes(namespace).Update(ctx, &binding, v1.UpdateOptions{})
			if err != nil {
				return knerrors.GetError(err)
			}
		} else {
			return fmt.Errorf("kamelet binding with name %q already exists. Use --force to recreate the binding", binding.Name)
		}
	}
	if existed {
		fmt.Printf("hhelo")
	}
	//if existed {
	//	_, _ = fmt.Fprintf(options.CmdOut, "kamelet binding %q updated\n", name)
	//} else {
	//	_, _ = fmt.Fprintf(options.CmdOut, "kamelet binding %q created\n", name)
	//}

	return nil
}

func nameFor(name, sink string, sourceRef corev1.ObjectReference) string {
	if name != "" {
		return name
	}

	generated := fmt.Sprintf("%s-to-%s-%s", sink, sourceRef.Kind, sourceRef.Name)

	generated = filepath.Base(generated)
	generated = strings.Split(generated, ".")[0]
	generated = strings.ToLower(generated)
	generated = disallowedChars.ReplaceAllString(generated, "")
	generated = strings.TrimFunc(generated, isDisallowedStartEndChar)

	return generated
}

func asEndpointProperties(props map[string]string) (camelv1.EndpointProperties, error) {
	if len(props) == 0 {
		return camelv1.EndpointProperties{}, nil
	}
	data, err := json.Marshal(props)
	if err != nil {
		return camelv1.EndpointProperties{}, err
	}
	return camelv1.EndpointProperties{
		RawMessage: camelv1.RawMessage(data),
	}, nil
}
func parseProperties(properties []string) (map[string]string, error) {
	props := make(map[string]string)
	for _, property := range properties {
		keyValue := strings.Split(property, "=")
		if len(keyValue) != 2 {
			return nil, fmt.Errorf("invalid property format: %s", property)
		}
		props[keyValue[0]] = keyValue[1]
	}
	return props, nil
}
func verifyProperties(kamelet *camelv1.Kamelet, endpoint camelv1.Endpoint) error {
	pMap, err := endpoint.Properties.GetPropertyMap()

	if kamelet.Spec.Definition != nil && len(kamelet.Spec.Definition.Required) > 0 {
		if err != nil {
			return err
		}
		for _, reqProp := range kamelet.Spec.Definition.Required {
			found := false
			if endpoint.Properties != nil {
				if _, contains := pMap[reqProp]; contains {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("binding is missing required property %q for Kamelet %q", reqProp, kamelet.Name)
			}
		}
	}

	for propName := range pMap {
		if _, ok := kamelet.Spec.Definition.Properties[propName]; !ok {
			return fmt.Errorf("binding uses unknown property %q for Kamelet %q", propName, kamelet.Name)
		}
	}

	return nil
}

func decodeSource(source string) (corev1.ObjectReference, error) {
	ref := corev1.ObjectReference{}

	if sourceExpression.MatchString(source) {
		groupNames := sourceExpression.SubexpNames()
		for _, match := range sourceExpression.FindAllStringSubmatch(source, -1) {
			for idx, text := range match {
				groupName := groupNames[idx]
				switch groupName {
				case "apiVersion":
					ref.APIVersion = text
				case "namespace":
					ref.Namespace = text
				case "kind":
					ref.Kind = text
				case "name":
					ref.Name = text
				}
			}
		}

		if sinkType, ok := sourceTypes[ref.Kind]; ok {
			if sinkType.Kind != "" {
				ref.Kind = sinkType.Kind
			}
			if ref.APIVersion == "" && sinkType.APIVersion != "" {
				ref.APIVersion = sinkType.APIVersion
			}
		} else {
			return ref, fmt.Errorf("unsupported sink type %q", ref.Kind)
		}
	} else {
		return ref, fmt.Errorf("unsupported sink expression %q - please use format <kind>:<name>", source)
	}

	return ref, nil
}

func getSourceProperties(options CreateBindingOptions) (map[string]string, error) {
	props := make(map[string]string)

	if options.SourceBrokerType != "" {
		props["type"] = options.SourceBrokerType
	}
	if options.CloudEventsSpecVersion != "" {
		props["cloudEventsSpecVersion"] = options.CloudEventsSpecVersion
	}

	if options.CloudEventsType != "" {
		props["cloudEventsType"] = options.CloudEventsType
	}

	overrideProps, err := parseProperties(options.CloudEventsOverride)
	if err != nil {
		return props, knerrors.GetError(err)
	}

	for key, prop := range overrideProps {
		props["ce.override."+key] = prop
	}

	return props, nil
}
