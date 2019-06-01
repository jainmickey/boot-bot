package utils

func Contains(searchable []interface{}, values []string) bool {
	for _, value := range values {
		for _, item := range searchable {
			if value == item {
				return true
			}
		}
	}
	return false
}
