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

type Server struct {
	podClient corev1.PodInterface
}

type postData struct {
	Warning  *string `json:"warning,omitempty"`
	Critical *string `json:"critical,omitempty"`
}

type CheckResponse struct {
	Code    *int32 `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (s *Server) ListenAndServe() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			var pd postData
			if err := json.Unmarshal(data, &pd); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			s.checkPodCount(w, &pd)

		default:
			http.Error(w, "", http.StatusForbidden)
			return
		}
	})
	if err := http.ListenAndServe(":80", http.DefaultServeMux); err != nil {
		log.Fatal(err)
	}
}

func (s *Server) checkPodCount(w http.ResponseWriter, pd *postData) {
	objects, err := s.podClient.List(metav1.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if pd.Critical != nil {
		cv, err := strconv.Atoi(*pd.Critical)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(objects.Items) >= cv {
			code := int32(2)
			resp := &CheckResponse{
				Code:    &code,
				Message: fmt.Sprintf(`More than "%d" pod exists`, cv),
			}
			data, err := json.Marshal(resp)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Write(data)
			return
		}
	}

	if pd.Warning != nil {
		cv, err := strconv.Atoi(*pd.Warning)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(objects.Items) >= cv {
			code := int32(1)
			resp := &CheckResponse{
				Code:    &code,
				Message: fmt.Sprintf(`More than "%d" pod exists`, cv),
			}
			data, err := json.Marshal(resp)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Write(data)
			return
		}
	}

	code := int32(0)
	resp := &CheckResponse{
		Code:    &code,
		Message: "",
	}
	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(data)
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

	opts := &Server{
		podClient: kClient.CoreV1().Pods(core.NamespaceAll),
	}

	opts.ListenAndServe()
}
