package main

import (
	"context"
	"encoding/json"
	"github.com/spf13/pflag"
	"io/ioutil"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/options"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/component-base/cli/globalflag"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	netPol = "labels-netpol"
)

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

type Options struct {
	SecureServingOptions options.SecureServingOptions
}

type Config struct {
	SecureServingInfo *server.SecureServingInfo
}

func NewDefaultOptions() *Options {
	opt := &Options{
		SecureServingOptions: *options.NewSecureServingOptions(),
	}
	opt.SecureServingOptions.BindPort = 8443
	opt.SecureServingOptions.ServerCert.PairName = netPol
	return opt
}

func (o *Options) GetConfig() *Config {
	err := o.SecureServingOptions.MaybeDefaultWithSelfSignedCerts("0.0.0.0", nil, nil)
	if err != nil {
		log.Fatalf("Error Getting Config.\nReason --> %s", err.Error())
	}
	c := Config{}
	err = o.SecureServingOptions.ApplyTo(&c.SecureServingInfo)
	if err != nil {
		return nil
	}
	return &c
}

func (o *Options) AddFlagSet(fs *pflag.FlagSet) {
	o.SecureServingOptions.AddFlags(fs)
}

func main() {
	defaultOptions := NewDefaultOptions()
	flagSet := pflag.NewFlagSet(netPol, pflag.ExitOnError)
	globalflag.AddGlobalFlags(flagSet, netPol)
	defaultOptions.AddFlagSet(flagSet)
	err := flagSet.Parse(os.Args)
	if err != nil {
		log.Fatalf("Not Able to Parse Flags.\nReason --> %s", err.Error())
	}
	c := defaultOptions.GetConfig()

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(ServeLabelValidation))

	stopCh := server.SetupSignalHandler()
	serve, _, err := c.SecureServingInfo.Serve(mux, 30*time.Second, stopCh)
	if err != nil {
		return
	} else {
		<-serve
	}

}

func ServeLabelValidation(writer http.ResponseWriter, request *http.Request) {
	log.Println("ServeLabelValidation was called")

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		responsewriters.InternalError(writer, request, err)
		log.Fatalf("Error Reading Body %s", err.Error())
	}
	gvk := admissionv1beta1.SchemeGroupVersion.WithKind("AdmissionReview")
	var admissionReview admissionv1beta1.AdmissionReview
	_, _, err = codecs.UniversalDeserializer().Decode(body, &gvk, &admissionReview)
	if err != nil {
		log.Fatalf("Error Converting Request Body into Admission Review Type %s", err.Error())
	}

	gvkPod := corev1.SchemeGroupVersion.WithKind("pods")
	var pod corev1.Pod
	_, _, err = codecs.UniversalDeserializer().Decode(admissionReview.Request.Object.Raw, &gvkPod, &pod)
	if err != nil {
		log.Fatalf("Error While getting pod type while admission review %s", err.Error())
	}
	log.Printf("Pod Resource we have is %v", pod)
	status := matchLabels(pod)
	var response admissionv1beta1.AdmissionResponse
	if status == false {
		log.Printf("Label Already Exists in Network Policy.....")
		response = admissionv1beta1.AdmissionResponse{
			UID:     admissionReview.Request.UID,
			Allowed: false,
			Result: &metav1.Status{
				Message: "Label Already Exists in Network Policy",
			},
		}
	} else {
		response = admissionv1beta1.AdmissionResponse{
			UID:     admissionReview.Request.UID,
			Allowed: true,
		}
	}
	admissionReview.Response = &response
	res, err := json.Marshal(admissionReview)
	if err != nil {
		log.Fatalf("Error Converting Response..")
	}
	_, err = writer.Write(res)
	if err != nil {
		return
	}

}

func matchLabels(pod corev1.Pod) bool {
	clientSet := getClientSet()
	netPolList, err := clientSet.NetworkingV1().NetworkPolicies(pod.Namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Printf("Error Occurred : %s\n while getting Network Policies in Namespace %s", err.Error(), pod.Namespace)
		return false
	}
	podLabels := pod.Labels

	netPolItems := netPolList.Items
	for _, netPolicy := range netPolItems {
		netPolSpecLabels := netPolicy.Spec.PodSelector.MatchLabels
		for podLabelKey, podLabelValue := range podLabels {
			for netPolSpecLabelKey, netPolSpecLabelValue := range netPolSpecLabels {
				if podLabelKey == netPolSpecLabelKey && podLabelValue == netPolSpecLabelValue {
					return false
				}
			}
		}

		netPolIngressPodSelector := new(map[string]string)
		netPolIngressRule := netPolicy.Spec.Ingress
		for _, val := range netPolIngressRule {
			for _, val1 := range val.From {
				netPolIngressPodSelector = &val1.PodSelector.MatchLabels
			}
		}

		for netPolIngressPodSelectorKey, netPolIngressPodSelectorValue := range *netPolIngressPodSelector {
			for podLabelKey, podLabelValue := range podLabels {
				if netPolIngressPodSelectorKey == podLabelKey && netPolIngressPodSelectorValue == podLabelValue {
					return false
				}
			}
		}

	}

	return true
}

func getClientSet() kubernetes.Interface {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("Not able to create kubeconfig object from inside pod.\nReason --> %s", err.Error())
	}
	log.Println("Created config object with In Cluster Config")

	// Creating Clientset
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error occurred while creating Client Set with provided config.\nReason --> %s", err.Error())
	}
	return clientSet
}
