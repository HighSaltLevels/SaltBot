package util

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Client interface for testing of 3rd party web services like youtube, giphy, jeopardy, etc...
type HttpClientInterface interface {
	Get(string) (*http.Response, error)
}

var unitDict map[string]int = map[string]int{
	"year":    31536000,
	"years":   31536000,
	"month":   2592000,
	"months":  2592000,
	"weeks":   604800,
	"week":    604800,
	"days":    86400,
	"day":     86400,
	"hours":   3600,
	"hour":    3600,
	"minutes": 60,
	"minute":  60,
	"seconds": 1,
	"second":  1,
}

// remove "-a" or "-i" from the args
func ParseArgsToQuery(args []string) string {
	parsedArgs := make([]string, 0)
	foundIndex := false
	for _, arg := range args {
		if foundIndex {
			foundIndex = false
			continue
		}

		if arg == "-i" {
			foundIndex = true
			continue
		}

		if arg == "-a" {
			continue
		}

		parsedArgs = append(parsedArgs, arg)
	}

	return strings.Join(parsedArgs, "+")
}

// Take in the unit of time and duration and return the unix epoch
func ParseExpiry(unit, duration string) (int64, error) {
	unitInt, ok := unitDict[unit]
	if !ok {
		return 0, fmt.Errorf("unparseable unit: %s", unit)
	}

	parsedDuration, err := strconv.Atoi(duration)
	if err != nil {
		return 0, fmt.Errorf("invalid duration: %w", err)
	}

	fullDuration := parsedDuration * unitInt
	expiry := time.Now().Unix() + int64(fullDuration)
	return expiry, nil
}

func TimeFromExpiry(expiry int64) string {
	expiryTime := time.Unix(expiry, 0)
	return expiryTime.Format(time.RFC1123)
}

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
