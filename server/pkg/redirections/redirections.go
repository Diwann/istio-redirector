package redirections

import (
	"bytes"
	"fmt"
	"istio-redirector/domain"
	"istio-redirector/pkg/csv"
	"istio-redirector/pkg/redirections/istio"
	"istio-redirector/utils"
	"net/url"
	"strconv"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Generate(inputData domain.InputData) (bytes.Buffer, error) {

	var istioConfig istio.Config
	var payload bytes.Buffer
	// Set the file name of the configurations file
	viper.SetConfigName("config")
	// Set the path to look for the configurations file
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.WithError(err).Info("can't read config.yaml")
	}

	err := viper.Unmarshal(&istioConfig)
	if err != nil {
		log.WithError(err).Info("unable to decode into struct")
	}

	r := istio.Redirections{
		Name:               inputData.RedirectionName,
		ClusterName:        inputData.RedirectionEnv,
		Namespace:          inputData.RedirectionNamespace,
		FallbackValueRegex: inputData.FallbackValueRegex,
		DestinationHost:    inputData.DestinationHost + ".svc.cluster.local",
		Gateways:           istioConfig.Istio.Gateways,
	}

	rulesCSV := csv.ReadFile(inputData.File)
	var hosts []string

	for _, rule := range rulesCSV {
		var data domain.Rule

		u, err := url.Parse(rule[0])
		if err != nil {
			panic(err)
		}

		hosts = append(hosts, u.Host)

		toRemoveFrom := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
		from := strings.ReplaceAll(rule[0], toRemoveFrom, "")

		if inputData.RedirectionType == "3xx" {
			code, err := strconv.Atoi(rule[2])
			if err != nil {
				log.WithError(err).Error("fail to parse line")
				break
			}

			dest, err := url.Parse(rule[1])
			if err != nil {
				log.WithError(err).Error("fail to parse line")
				break
			}
			toRemoveTo := fmt.Sprintf("%s://%s", dest.Scheme, dest.Host)
			to := strings.ReplaceAll(rule[1], toRemoveTo, "")
			if to == "" {
				to = "/"
			}

			data = domain.Rule{
				From: from,
				To:   to,
				Code: code,
			}
		} else {
			code, _ := strconv.Atoi(rule[1])
			data = domain.Rule{
				From: from,
				Code: code,
			}
		}

		r.Rules = append(r.Rules, data)
	}

	r.Hosts = utils.RemoveDuplicates(hosts)
	if len(r.Hosts) > 1 {
		return payload, fmt.Errorf("found %v hosts in the source file, should be 1", len(r.Hosts))
	}

	t, err := template.ParseFiles("templates/virtual-service.yaml")
	if err != nil {
		log.WithError(err).Error("fail to parse template")
		return payload, err
	}

	err = t.Execute(&payload, r)
	if err != nil {
		log.WithError(err).Error("fail to execute content to template")
		return payload, err
	}

	_, err = istio.Validate(&payload)
	if err != nil {
		log.WithError(err).Error("fail to validate template as VirtualService")
		return payload, err
	}

	return payload, nil
}
