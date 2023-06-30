package v1

import (
	"fmt"
	"time"
)

// Time3339 is a time.Time which encodes to and from JSON
// as an RFC 3339 time in UTC.
// Copied from golang.org/src/time/time.go
type Time3339 time.Time

func (t *Time3339) UnmarshalJSON(b []byte) error {
	if len(b) < 2 || b[0] != '"' || b[len(b)-1] != '"' {
		return fmt.Errorf("failed to unmarshal non-string value %q as an RFC 3339 time", b)
	}
	tm, err := time.Parse(time.RFC3339, string(b[1:len(b)-1]))
	if err != nil {
		return err
	}
	*t = Time3339(tm)
	return nil
}

func (t Time3339) MarshalJSON() ([]byte, error) {
	tm := time.Time(t)
	b := make([]byte, 0, len(time.RFC3339)+2)
	b = append(b, '"')
	b = tm.AppendFormat(b, time.RFC3339)
	b = append(b, '"')
	return b, nil
}
