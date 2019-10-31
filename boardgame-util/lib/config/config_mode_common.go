package config

import (
	"strings"
)

//ModeField denotes the field in a RawConfigMode. Used to pass to the
//UpdateConfig family of function factories.
type ModeField string

const (
	//FieldInvalid is an invalid default value
	FieldInvalid ModeField = "<INVALID>"
	//FieldAllowedOrigins denotes that field of the output
	FieldAllowedOrigins = "AllowedOrigins"
	//FieldDefaultPort denotes that field of the output
	FieldDefaultPort = "DefaultPort"
	//FieldDefaultStaticPort denotes that field of the output
	FieldDefaultStaticPort = "DefaultStaticPort"
	//FieldAdminUserIds denotes that field of the output
	FieldAdminUserIds = "AdminUserIds"
	//FieldDisableAdminChecking denotes that field of the output
	FieldDisableAdminChecking = "DisableAdminChecking"
	//FieldOfflineDevMode denotes that field of the output
	FieldOfflineDevMode = "OfflineDevMode"
	//FieldStorage denotes that field of the output
	FieldStorage = "Storage"
	//FieldDefaultStorageType denotes that field of the output
	FieldDefaultStorageType = "DefaultStorageType"
	//FieldGoogleAnalytics denotes that field of the output
	FieldGoogleAnalytics = "GoogleAnalytics"
	//FieldFirebase denotes that field of the output
	FieldFirebase = "Firebase"
	//FieldAPIHost denotes that field of the output
	FieldAPIHost = "ApiHost"
	//FieldGames denotes that field of the output
	FieldGames = "Games"
)

//ModeFieldType is the type of field in RawConfigMode. Different UpdateConfig
//function factories only work on certain types of fields.
type ModeFieldType int

const (
	//FieldTypeInvalid is an invalid default value
	FieldTypeInvalid ModeFieldType = iota
	//FieldTypeString denotes a field with type string
	FieldTypeString
	//FieldTypeStringSlice denotes a field with type []string
	FieldTypeStringSlice
	//FieldTypeStringMap denotes a field with type map[string]string
	FieldTypeStringMap
	//FieldTypeBool denotes a field with type bool
	FieldTypeBool
	//FieldTypeFirebase denotes a field with type firebase
	FieldTypeFirebase
	//FieldTypeGameNode denotes a field with type GameNode
	FieldTypeGameNode
)

//FieldTypes maps each ConfigModeField to its ConfigModeFieldType.
var FieldTypes = map[ModeField]ModeFieldType{
	FieldInvalid:              FieldTypeInvalid,
	FieldAllowedOrigins:       FieldTypeString,
	FieldDefaultPort:          FieldTypeString,
	FieldDefaultStaticPort:    FieldTypeString,
	FieldAdminUserIds:         FieldTypeStringSlice,
	FieldDisableAdminChecking: FieldTypeBool,
	FieldOfflineDevMode:       FieldTypeBool,
	FieldStorage:              FieldTypeStringMap,
	FieldDefaultStorageType:   FieldTypeString,
	FieldGoogleAnalytics:      FieldTypeString,
	FieldFirebase:             FieldTypeFirebase,
	FieldAPIHost:              FieldTypeString,
	FieldGames:                FieldTypeGameNode,
}

//ModeCommon is the values that both ConfigMode and RawConfigMode share
//directly, factored out for convenience so they can be anonymously embedded in
//ConfigMdoe and RawConfigMode.
type ModeCommon struct {
	AllowedOrigins    string   `json:"allowedOrigins,omitempty"`
	DefaultPort       string   `json:"defaultPort,omitempty"`
	DefaultStaticPort string   `json:"defaultStaticPort,omitempty"`
	AdminUserIds      []string `json:"adminUserIds,omitempty"`
	//This is a dangerous config. Only enable in Dev!
	DisableAdminChecking bool `json:"disableAdminChecking,omitempty"`
	//This is a dangerous config, designed to only be used in dev.
	OfflineDevMode bool              `json:"offlineDevMode,omitempty"`
	Storage        map[string]string `json:"storage,omitempty"`
	//The storage type that should be used if no storage type is provided via
	//command line options.
	DefaultStorageType string `json:"defaultStorageType,omitempty"`

	//The GA config string. Will be used to generate the client_config json
	//blob. Generally has a structure like "UA-321655-11"
	GoogleAnalytics string          `json:"googleAnalytics,omitempty"`
	Firebase        *FirebaseConfig `json:"firebase,omitempty"`
	//The host name the client should connect to in that mode. Something like
	//"http://localhost:8888"
	APIHost string `json:"apiHost,omitempty"`
}

//FieldFromString returns a ModeField by doing fuzzing matching.
func FieldFromString(s string) ModeField {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	for key := range FieldTypes {
		normalizedKey := strings.ToLower(string(key))

		if normalizedKey == s {
			return key
		}
	}

	return FieldInvalid
}
