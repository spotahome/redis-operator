package k8s

import (
	"crypto/sha256"
	"encoding/base64"
	"hash"

	"github.com/davecgh/go-spew/spew"
)

type Annotated interface {
	GetAnnotations() map[string]string
	SetAnnotations(annotations map[string]string)
	GetName() string
}

const resourceHashAnnotationKey = "databases.spotahome.com/resource-hash"

// Create hash of a given object

func addHashAnnotation(r Annotated) {
	hash := deepHashString(r)
	m := r.GetAnnotations()
	if m == nil {
		m = map[string]string{}
	}
	m[resourceHashAnnotationKey] = hash
	r.SetAnnotations(m)
}

func deepHashString(obj interface{}) string {
	hasher := sha256.New()
	deepHashObject(hasher, obj)
	hashBytes := hasher.Sum([]byte{})
	b64Hash := base64.StdEncoding.EncodeToString(hashBytes)
	return b64Hash
}

// DeepHashObject writes specified object to hash using the spew library
// which follows pointers and prints actual values of the nested objects
// ensuring the hash does not change when a pointer changes.
func deepHashObject(hasher hash.Hash, objectToWrite interface{}) {
	hasher.Reset()
	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	printer.Fprintf(hasher, "%#v", objectToWrite)
}

func shouldUpdate(desired Annotated, stored Annotated) bool {

	storedHash, exists := stored.GetAnnotations()[resourceHashAnnotationKey]
	if !exists {
		return true
	}
	desiredHash := deepHashString(desired)

	return desiredHash != storedHash
}
