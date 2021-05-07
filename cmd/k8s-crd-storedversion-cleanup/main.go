package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"strings"

	apiextensionsV1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	errNoStorageVersionName = "no storage version name found"
)

var (
	apiGroup string
	dryRun   bool
	help     bool
)

func init() {
	flag.StringVar(&apiGroup, "group", "", "the spec.group that should be filtered. Matches to the suffix so api.io goes for a.api.io and b.api.io")
	flag.BoolVar(&dryRun, "dry-run", false, "kubectl dry-run")
	flag.BoolVar(&help, "help", false, "show help")
}

func getStorageVersionName(crd *apiextensionsV1.CustomResourceDefinition) string {
	for _, version := range crd.Spec.Versions {
		if version.Storage {
			return version.Name
		}
	}

	return ""
}

func cleanupStorageVersion(crd *apiextensionsV1.CustomResourceDefinition) error {
	storageVersionName := getStorageVersionName(crd)
	if storageVersionName == "" {
		return errors.New(errNoStorageVersionName)
	}

	crd.Status.StoredVersions = []string{storageVersionName}
	return nil
}

type crdFilter func(*apiextensionsV1.CustomResourceDefinition) bool

func filterMatchAll() crdFilter {
	return func(_ *apiextensionsV1.CustomResourceDefinition) bool {
		return true
	}
}

func filterMatchGroupSuffix(suffix string) crdFilter {
	return func(crd *apiextensionsV1.CustomResourceDefinition) bool {
		return strings.HasSuffix(crd.Spec.Group, suffix)
	}
}

func cleanupCrds(filter crdFilter) {
	//  create k8s client
	cfg := config.GetConfigOrDie()
	client, err := apiextension.NewForConfig(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	// retrieve CRD
	crds, err := client.ApiextensionsV1().CustomResourceDefinitions().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		log.Fatalln(err)
	}

	opts := v1.UpdateOptions{}
	if dryRun {
		opts.DryRun = []string{"All"}
	}

	for _, crd := range crds.Items {
		if !filter(&crd) {
			continue
		}

		if err := cleanupStorageVersion(&crd); err != nil {
			log.Printf("%s: %s", crd.ObjectMeta.Name, err)
		}

		crd, err := client.ApiextensionsV1().CustomResourceDefinitions().UpdateStatus(context.TODO(), &crd, opts)
		if err != nil {
			log.Printf("%s: %s", crd.ObjectMeta.Name, err)
		} else {
			log.Printf("%s: updated status.storedVersion: %s", crd.ObjectMeta.Name, crd.Status.StoredVersions)
		}
	}
}

func main() {
	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
	}

	var filter crdFilter
	switch apiGroup {
	case "":
		filter = filterMatchAll()
	default:
		filter = filterMatchGroupSuffix(apiGroup)
	}

	cleanupCrds(filter)
}
