package naming

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"unicode"
)

// ToValidImageName converts the given string into a valid docker image name
func ToValidImageName(name string) string {
	return strings.ToLower(name)
}

// ToValidImageVersion converts the given string into a valid docker image version string
func ToValidImageVersion(version string) string {
	return strings.Replace(version, ":", "", -1)
}

// ToValidName converts the given string into a valid Kubernetes resource name
func ToValidName(name string) string {
	return toValidName(name, false, math.MaxInt32)
}

// ToValidNameWithDots converts the given string into a valid Kubernetes resource name
func ToValidNameWithDots(name string) string {
	return toValidName(name, true, math.MaxInt32)
}

// ToValidNameTruncated converts the given string into a valid Kubernetes resource name,
// truncating the result if it is more than maxLength characters.
func ToValidNameTruncated(name string, maxLength int) string {
	return toValidName(name, false, maxLength)
}

func toValidName(name string, allowDots bool, maxLength int) string {
	if name == "" {
		return ""
	}
	var buffer bytes.Buffer
	first := true
	lastCharDash := false
	hasLetter := false
	for _, ch := range name {
		ch = unicode.ToLower(ch)
		if ch >= 'a' && ch <= 'z' {
			hasLetter = true
			break
		}
	}
	if !hasLetter {
		name = fmt.Sprintf("x%s", name)
	}
	for _, ch := range name {
		ch = unicode.ToLower(ch)
		if buffer.Len()+1 > maxLength {
			break
		}
		if first {
			// strip non letters at start
			if ch >= 'a' && ch <= 'z' {
				buffer.WriteRune(ch)
				first = false
			}
		} else {
			if !allowDots && ch == '.' {
				ch = '-'
			}
			if !(ch >= 'a' && ch <= 'z') && !(ch >= '0' && ch <= '9') && ch != '-' && ch != '.' {
				ch = '-'
			}

			if ch != '-' || !lastCharDash {
				buffer.WriteRune(ch)
			}
			lastCharDash = ch == '-'
		}
	}
	answer := buffer.String()
	for {
		if strings.HasSuffix(answer, "-") {
			answer = strings.TrimSuffix(answer, "-")
		} else {
			break
		}
	}
	return answer
}

//EmailToK8sID converts the provided email address to a valid Kubernetes resource name, converting the @ to a .
func EmailToK8sID(email string) string {
	return ToValidNameWithDots(strings.Replace(email, "@", ".", -1))
}
