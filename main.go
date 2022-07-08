package main

import (
	"github.com/spf13/pflag"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/options"
	"k8s.io/component-base/cli/globalflag"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	netPol = "labels-netpol"
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
}
