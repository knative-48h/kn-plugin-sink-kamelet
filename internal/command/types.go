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
	"io"

	camelk "github.com/apache/camel-k/v2/pkg/client/camel/clientset/versioned"
	camelv1 "github.com/apache/camel-k/v2/pkg/client/camel/clientset/versioned/typed/camel/v1"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/client-pkg/pkg/kn/commands"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

const (
	KameletTypeLabel              = "camel.apache.org/kamelet.type"
	KameletSupportLevelAnnotation = "camel.apache.org/kamelet.support.level"
	KameletProviderAnnotation     = "camel.apache.org/provider"
)

var (
	sourceTypes = map[string]corev1.ObjectReference{
		"channel": {
			Kind:       "Channel",
			APIVersion: messagingv1.SchemeGroupVersion.String(),
		},
		"broker": {
			Kind:       "Broker",
			APIVersion: eventingv1.SchemeGroupVersion.String(),
		},
		"ksvc": {
			Kind:       "Service",
			APIVersion: servingv1.SchemeGroupVersion.String(),
		},
	}
)

type KameletPluginParams struct {
	*commands.KnParams
	Context          context.Context
	ContextCancel    context.CancelFunc
	NewKameletClient func() (camelv1.CamelV1Interface, error)
}

func (params *KameletPluginParams) Initialize() {
	if params.KnParams == nil {
		params.KnParams = &commands.KnParams{}
		params.KnParams.Initialize()
	}

	if params.NewKameletClient == nil {
		params.NewKameletClient = params.newKameletClient
	}
}

func (params *KameletPluginParams) newKameletClient() (camelv1.CamelV1Interface, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	client, err := camelk.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return client.CamelV1(), nil
}

// CreateBindingOptions holding settings and options on the create binding command
type CreateBindingOptions struct {
	Name                   string
	Sink                   string
	SinkProperties         []string
	CloudEventsOverride    []string
	CloudEventsSpecVersion string
	CloudEventsType        string
	Source                 string
	SourceBrokerType       string
	Broker                 string
	Channel                string
	Service                string
	Force                  bool
	CmdOut                 io.Writer
}
