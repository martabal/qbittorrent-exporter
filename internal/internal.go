package internal

import (
	"net/url"
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
		*input = "/" + *input
	}
}

func CompareSemVer(v1, v2 string) int {
	i1, i2 := 0, 0

	for i1 < len(v1) || i2 < len(v2) {
		n1 := 0

		for i1 < len(v1) && v1[i1] != '.' {
			if v1[i1] >= '0' && v1[i1] <= '9' {
				n1 = n1*10 + int(v1[i1]-'0')
			}

			i1++
		}

		if i1 < len(v1) {
			i1++
		}

		n2 := 0

		for i2 < len(v2) && v2[i2] != '.' {
			if v2[i2] >= '0' && v2[i2] <= '9' {
				n2 = n2*10 + int(v2[i2]-'0')
			}

			i2++
		}

		if i2 < len(v2) {
			i2++
		}

		if n1 < n2 {
			return -1
		} else if n1 > n2 {
			return 1
		}
	}

	return 0
}
