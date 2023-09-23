package util

// MergeLabels merges all the label maps received as argument into a single new label map.
func MergeLabels(allLabels ...map[string]string) map[string]string {
	res := map[string]string{}

	for _, labels := range allLabels {
		for k, v := range labels {
			res[k] = v
		}
	}
	return res
}

// MergeAnnotations merges all the annotations maps received as argument into a single new label map.
func MergeAnnotations(allMergeAnnotations ...map[string]string) map[string]string {
	res := map[string]string{}

	for _, annotations := range allMergeAnnotations {
		for k, v := range annotations {
			res[k] = v
		}
	}
	return res
}
