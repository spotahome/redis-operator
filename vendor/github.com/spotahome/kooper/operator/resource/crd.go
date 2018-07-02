package resource

import (
	"github.com/spotahome/kooper/operator/retrieve"
)

// CRD represents a non stadandard resource or custom resource definition.
type CRD interface {
	retrieve.Retriever
	// Initialize knows how to ensure that the CRD is initialized.
	Initialize() error
}
