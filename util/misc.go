package util

type RemoteFileFetcher func(url string) ([]byte, error)

func MapKeys(source map[string]bool) []string {
	values := make([]string, 0, 0)
	for key, _ := range source {
		values = append(values, key)
	}
	return values
}
