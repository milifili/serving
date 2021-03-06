/*
Copyright 2018 The Knative Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/metrics"
	"knative.dev/pkg/system"

	. "knative.dev/pkg/configmap/testing"
	_ "knative.dev/pkg/system/testing"
)

func TestOurObservability(t *testing.T) {
	cm, example := ConfigMapsFromTestFile(t, metrics.ConfigMapName())

	realCfg, err := metrics.NewObservabilityConfigFromConfigMap(cm)
	if err != nil {
		t.Fatal("NewObservabilityConfigFromConfigMap(actual) =", err)
	}
	if realCfg == nil {
		t.Fatal("NewObservabilityConfigFromConfigMap(actual) = nil")
	}

	exCfg, err := metrics.NewObservabilityConfigFromConfigMap(example)
	if err != nil {
		t.Fatal("NewObservabilityConfigFromConfigMap(example) =", err)
	}
	if exCfg == nil {
		t.Fatal("NewObservabilityConfigFromConfigMap(example) = nil")
	}
	// TODO(#8644): remove this.
	co := cmpopts.IgnoreFields(metrics.ObservabilityConfig{}, "RequestLogTemplate")
	if !cmp.Equal(realCfg, exCfg, co) {
		t.Errorf("actual != example: diff(-actual,+exCfg):\n%s", cmp.Diff(realCfg, exCfg, co))
	}
}

func TestObservabilityConfiguration(t *testing.T) {
	observabilityConfigTests := []struct {
		name           string
		wantErr        bool
		wantController interface{}
		config         *corev1.ConfigMap
	}{{
		name:    "observability configuration with all inputs",
		wantErr: false,
		wantController: &metrics.ObservabilityConfig{
			LoggingURLTemplate:     "https://logging.io",
			EnableVarLogCollection: true,
			RequestLogTemplate:     `{"requestMethod": "{{.Request.Method}}"}`,
			EnableRequestLog:       true,
			EnableProbeRequestLog:  true,
			RequestMetricsBackend:  "stackdriver",
		},
		config: &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: system.Namespace(),
				Name:      metrics.ConfigMapName(),
			},
			Data: map[string]string{
				"logging.enable-var-log-collection":           "true",
				"logging.revision-url-template":               "https://logging.io",
				"logging.enable-probe-request-log":            "true",
				"logging.write-request-logs":                  "true",
				"logging.request-log-template":                `{"requestMethod": "{{.Request.Method}}"}`,
				"metrics.request-metrics-backend-destination": "stackdriver",
			},
		},
	}, {
		name:    "observability config with no map",
		wantErr: false,
		wantController: &metrics.ObservabilityConfig{
			EnableVarLogCollection: false,
			LoggingURLTemplate:     metrics.DefaultLogURLTemplate,
			RequestLogTemplate:     "",
			RequestMetricsBackend:  "prometheus",
		},
		config: &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: system.Namespace(),
				Name:      metrics.ConfigMapName(),
			},
		},
	}, {
		name:           "invalid request log template",
		wantErr:        true,
		wantController: (*metrics.ObservabilityConfig)(nil),
		config: &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: system.Namespace(),
				Name:      metrics.ConfigMapName(),
			},
			Data: map[string]string{
				"logging.request-log-template": `{{ something }}`,
			},
		},
	}}

	for _, tt := range observabilityConfigTests {
		t.Run(tt.name, func(t *testing.T) {
			actualController, err := metrics.NewObservabilityConfigFromConfigMap(tt.config)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Test: %q; NewObservabilityFromConfigMap() error = %v, WantErr %v", tt.name, err, tt.wantErr)
			}

			if diff := cmp.Diff(actualController, tt.wantController); diff != "" {
				t.Fatalf("Test: %q; want %v, but got %v", tt.name, tt.wantController, actualController)
			}
		})
	}
}
