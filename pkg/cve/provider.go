package cve

import (
	"github.com/jenkins-x/jx/pkg/client/clientset/versioned"
	"github.com/jenkins-x/jx/pkg/table"
	"k8s.io/client-go/kubernetes"
)

const (
	AnnotationCVEImageId = "jenkins-x.io/cve-image-id"
)

type CVEQuery struct {
	ImageName       string
	ImageID         string
	Vesion          string
	Environment     string
	TargetNamespace string
}
type CVEProvider interface {
	GetImageVulnerabilityTable(jxClient versioned.Interface, client kubernetes.Interface, table *table.Table, query CVEQuery) error
}
