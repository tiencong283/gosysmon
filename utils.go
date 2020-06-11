package main

import "strings"

type StringSet map[string]struct{}

func NewStringSet() StringSet {
	return make(map[string]struct{})
}
func (set StringSet) Add(s string) {
	set[s] = struct{}{}
}
func (set StringSet) Has(s string) bool {
	if _, ok := set[s]; !ok {
		return false
	}
	return true
}
func (set StringSet) AddFromSet(entries StringSet) {
	for e := range entries {
		set.Add(e)
	}
}

// ToJson serializes the object into json format
func ToJson(v interface{}) string {
	bytes, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return err.Error()
	}
	return string(bytes)
}

// StringToMap converts the string in format: key1=value1,key2=value2,... to the map
func StringToMap(s string) map[string]string {
	tokens := strings.Split(s, ",")
	m := make(map[string]string, len(tokens))
	for _, token := range tokens {
		equalAt := strings.IndexByte(token, '=')
		if equalAt == -1 || equalAt == 0 || equalAt == len(token)-1 {
			continue
		}
		m[token[:equalAt]] = token[equalAt+1:]
	}
	return m
}

// GetImageName returns the executable name of the path
func GetImageName(path string) string {
	if path == "" {
		return ""
	}
	i := len(path) - 1
	for i >= 0 && path[i] != '\\' {
		i--
	}
	if i >= 0 {
		path = path[i+1:]
	}
	return path
}

// copy from https://github.com/golang/go/blob/master/src/path/filepath/path_windows.go

func isSlash(c uint8) bool {
	return c == '\\' || c == '/'
}

// reservedNames lists reserved Windows names. Search for PRN in
// https://docs.microsoft.com/en-us/windows/desktop/fileio/naming-a-file
// for details.
var reservedNames = []string{
	"CON", "PRN", "AUX", "NUL",
	"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
	"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
}

// isReservedName returns true, if path is Windows reserved name.
// See reservedNames for the full list.
func isReservedName(path string) bool {
	if len(path) == 0 {
		return false
	}
	for _, reserved := range reservedNames {
		if strings.EqualFold(path, reserved) {
			return true
		}
	}
	return false
}

// volumeNameLen returns length of the leading volume name on Windows.
// It returns 0 elsewhere.
func volumeNameLen(path string) int {
	if len(path) < 2 {
		return 0
	}
	// with drive letter
	c := path[0]
	if path[1] == ':' && ('a' <= c && c <= 'z' || 'A' <= c && c <= 'Z') {
		return 2
	}
	// is it UNC? https://msdn.microsoft.com/en-us/library/windows/desktop/aa365247(v=vs.85).aspx
	if l := len(path); l >= 5 && isSlash(path[0]) && isSlash(path[1]) &&
		!isSlash(path[2]) && path[2] != '.' {
		// first, leading `\\` and next shouldn't be `\`. its server name.
		for n := 3; n < l-1; n++ {
			// second, next '\' shouldn't be repeated.
			if isSlash(path[n]) {
				n++
				// third, following something characters. its share name.
				if !isSlash(path[n]) {
					if path[n] == '.' {
						break
					}
					for ; n < l; n++ {
						if isSlash(path[n]) {
							break
						}
					}
					return n
				}
				break
			}
		}
	}
	return 0
}

// WindowsIsAbs reports whether the path is absolute.
func WindowsIsAbs(path string) (b bool) {
	if isReservedName(path) {
		return true
	}
	l := volumeNameLen(path)
	if l == 0 {
		return false
	}
	path = path[l:]
	if path == "" {
		return false
	}
	return isSlash(path[0])
}
