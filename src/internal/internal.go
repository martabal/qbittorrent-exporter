package internal

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func IsValidURL(input string) bool {
	u, err := url.Parse(input)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func IsValidHttpsURL(input string) bool {
	u, err := url.Parse(input)
	return err == nil && u.Scheme == "https" && u.Host != ""
}

func EnsureLeadingSlash(input *string) {
	if !strings.HasPrefix(*input, "/") {
		*input = fmt.Sprintf("/%s", *input)
	}
}

func CompareSemVer(v1, v2 *string) int {
	v1Parts := strings.Split(*v1, ".")
	v2Parts := strings.Split(*v2, ".")

	for i := 0; i < len(v1Parts) || i < len(v2Parts); i++ {
		var n1 int
		if i < len(v1Parts) {
			n1, _ = strconv.Atoi(v1Parts[i])
		}

		var n2 int
		if i < len(v2Parts) {
			n2, _ = strconv.Atoi(v2Parts[i])
		}

		if n1 < n2 {
			return -1
		} else if n1 > n2 {
			return 1
		}
	}

	return 0
}
