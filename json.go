package boardgame

import (
	"encoding/json"
)

//A JSONObject is an object that is ready to be serialized to JSON.
type JSONObject interface{}

//JSONMap is just convenient where you need to create a map
type JSONMap map[string]JSONObject

type JSONer interface {
	//Returns the canonical JSON representation of this object, suitable to
	//being communicated across the wire or saved in a DB.
	JSON() JSONObject
}

//Serialize converts the JSONObject into bytes, suitable for being transferred
//across the wire or written to disk.
func Serialize(o JSONObject) []byte {
	result, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return []byte("")
	}
	return result
}
