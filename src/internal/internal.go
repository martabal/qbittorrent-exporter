package internal

import (
	"fmt"
	"net/url"
	"strings"
)

func IsValidURL(input string) bool {
	u, err := url.Parse(input)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func EnsureLeadingSlash(input *string) {
	if !strings.HasPrefix(*input, "/") {
		*input = fmt.Sprintf("/%s", *input)
	}
}
