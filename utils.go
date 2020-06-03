package main

// ToJson serializes the object into json format
func ToJson(v interface{}) string {
	bytes, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return err.Error()
	}
	return string(bytes)
}