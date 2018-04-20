package pkg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type options struct {
	podClient corev1.PodInterface
}

func NewCmdServer() *cobra.Command {
	cmd := &cobra.Command{
		Use: "run",
		Run: func(cmd *cobra.Command, args []string) {

			config, err := rest.InClusterConfig()
			if err != nil {
				log.Fatal(err)
			}

			kClient, err := kubernetes.NewForConfig(config)
			if err != nil {
				log.Fatal(err)
			}

			opts := &options{
				podClient: kClient.CoreV1().Pods(core.NamespaceAll),
			}

			opts.ListenAndServe()
		},
	}

	return cmd
}

type postData struct {
	Warning  *string `json:"warning,omitempty"`
	Critical *string `json:"critical,omitempty"`
}

func (opts *options) ListenAndServe() {

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

			opts.checkPodCount(w, &pd)

		default:
			http.Error(w, "", http.StatusForbidden)
			return
		}
	})

	s := &http.Server{
		Addr:           ":80",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

type WebhookResp struct {
	Code    *int32      `json:"code,omitempty"`
	Message interface{} `json:"message,omitempty"`
}

func (opts *options) checkPodCount(w http.ResponseWriter, pd *postData) {

	list, err := opts.podClient.List(meta_v1.ListOptions{})
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

		if len(list.Items) >= cv {
			code := int32(2)
			resp := &WebhookResp{
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

		if len(list.Items) >= cv {
			code := int32(1)
			resp := &WebhookResp{
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
	resp := &WebhookResp{
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
