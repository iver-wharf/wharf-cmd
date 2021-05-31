package namespace

import (
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Namespaces struct {
	Kubeconfig *rest.Config
}

func (n Namespaces) SetupNamespaces(namespaces []string) error {
	log.Traceln("SetupNamespace called")

	clientSet, err := kubernetes.NewForConfig(n.Kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	nsClient := clientSet.CoreV1().Namespaces()
	for _, name := range namespaces {
		ns, err := nsClient.Get(name, metav1.GetOptions{})

		if ns.Name == "" {
			log.WithField("name", name).Infoln("Creating namespace")

			ns.Name = name
			_, err = nsClient.Create(ns)
			if err != nil {
				log.WithError(err).Errorln("Error creating namespace")
			}

			continue
		}

		log.WithField("name", name).Traceln("Namespace found!")
	}

	return nil
}
