package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type Plugin struct {
	client corev1.PodInterface
}

type Request struct {
	Warning  *string `json:"warning,omitempty"`
	Critical *string `json:"critical,omitempty"`
}

type State int32

const (
	OK       State = iota // 0
	Warning               // 1
	Critical              // 2
	Unknown               // 3
)

type Response struct {
	Code    State  `json:"code"`
	Message string `json:"message,omitempty"`
}

func (p *Plugin) Check(req *Request) (*Response, error) {
	objects, err := p.client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	if req.Critical != nil {
		cv, err := strconv.Atoi(*req.Critical)
		if err != nil {
			return nil, err
		}

		if len(objects.Items) >= cv {
			return &Response{
				Code:    Critical,
				Message: fmt.Sprintf(`More than "%d" pod exists`, cv),
			}, nil
		}
	}

	if req.Warning != nil {
		cv, err := strconv.Atoi(*req.Warning)
		if err != nil {
			return nil, err
		}

		if len(objects.Items) >= cv {
			return &Response{
				Code:    Warning,
				Message: fmt.Sprintf(`More than "%d" pod exists`, cv),
			}, nil
		}
	}
	return &Response{Code: OK}, nil
}

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	kClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	p := &Plugin{
		client: kClient.CoreV1().Pods(core.NamespaceAll),
	}

	http.HandleFunc("/check-pod-count", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			var req Request
			if err := json.Unmarshal(data, &req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			resp, err := p.Check(&req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			err = json.NewEncoder(w).Encode(resp)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		default:
			http.Error(w, "", http.StatusNotImplemented)
			return
		}
	})
	http.ListenAndServe(":80", http.DefaultServeMux)
}
