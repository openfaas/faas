package middleware

import "strings"

func GetNamespace(defaultNamespace, fullName string) (string, string) {
	if index := strings.LastIndex(fullName, "."); index > -1 {
		return fullName[:index], fullName[index+1:]
	}
	return fullName, defaultNamespace
}
