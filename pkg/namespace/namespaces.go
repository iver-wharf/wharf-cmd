package namespace

import (
	"context"

	"github.com/iver-wharf/wharf-core/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var log = logger.New()

type Namespaces struct {
	Kubeconfig *rest.Config
}

func (n Namespaces) SetupNamespaces(namespaces []string) error {
	log.Debug().Message("SetupNamespace called")

	clientSet, err := kubernetes.NewForConfig(n.Kubeconfig)
	if err != nil {
		log.Panic().WithError(err).Message("Failed loading kube-config.")
	}

	nsClient := clientSet.CoreV1().Namespaces()
	for _, name := range namespaces {
		ns, err := nsClient.Get(context.TODO(), name, metav1.GetOptions{})

		if ns.Name == "" {
			log.Debug().WithString("namespace", name).Message("Creating namespace.")

			ns.Name = name
			_, err = nsClient.Create(context.TODO(), ns, metav1.CreateOptions{})
			if err != nil {
				log.Error().WithError(err).Message("Failed to create namespace.")
			} else {
				log.Info().WithString("namespace", name).Message("Created namespace.")
			}

			continue
		}

		log.Info().WithString("namespace", name).Message("Namespace already exists.")
	}

	return nil
}
