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

	maxLen := max(len(v1Parts), len(v2Parts))
	for len(v1Parts) < maxLen {
		v1Parts = append(v1Parts, "0")
	}
	for len(v2Parts) < maxLen {
		v2Parts = append(v2Parts, "0")
	}

	for i := 0; i < maxLen; i++ {
		n1, _ := strconv.Atoi(v1Parts[i])
		n2, _ := strconv.Atoi(v2Parts[i])

		if n1 < n2 {
			return -1
		} else if n1 > n2 {
			return 1
		}
	}

	return 0
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
