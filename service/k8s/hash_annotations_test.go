package k8s

import "testing"

// create test for addHashAnnotation

// create a dummy struct that implements Annotated interface
type dummy struct {
	annotations map[string]string
	name        string
}

func (d *dummy) GetAnnotations() map[string]string {
	return d.annotations
}

func (d *dummy) SetAnnotations(annotations map[string]string) {
	d.annotations = annotations
}

func (d *dummy) GetName() string {
	return d.name
}

func TestAddHashAnnotation(t *testing.T) {
	originalObject := &dummy{name: "test"}
	copyOfOriginalObject := &dummy{name: "test"}
	differentObject := &dummy{name: "test2"}

	addHashAnnotation(originalObject)

	tests := []struct {
		name         string
		object       Annotated
		errorMessage string
		expected     bool
	}{
		{
			name:         "Hashes of same object should be equal",
			object:       copyOfOriginalObject,
			errorMessage: "Hashes of same object should be equal",
			expected:     true,
		},
		{
			name:         "Hashes of different objects should not be equal",
			object:       differentObject,
			errorMessage: "Hashes of different objects should not be equal",
			expected:     false,
		},
	}
	for _, test := range tests {
		addHashAnnotation(test.object)
		hash := test.object.GetAnnotations()[resourceHashAnnotationKey]
		if hash == "" {
			t.Errorf("Hash not created")
		}
		equal := hash == originalObject.GetAnnotations()[resourceHashAnnotationKey]
		if equal != test.expected {
			t.Error(test.errorMessage)
		}
	}
}
