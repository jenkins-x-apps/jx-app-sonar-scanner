package util

import "encoding/json"

// PrettyPrint returns an indented string representation of the passed struct for the purpose of logging/debugging.
func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
