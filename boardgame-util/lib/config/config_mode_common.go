package config

//ConfigModeField denotes the field in a RawConfigMode. Used to pass to the
//UpdateConfig family of function factories.
type ConfigModeField string

const (
	FieldAllowedOrigins       ConfigModeField = "AllowedOrigins"
	FieldDefaultPort                          = "DefaultPort"
	FieldDefaultStaticPort                    = "DefaultStaticPort"
	FieldAdminUserIds                         = "AdminUserIds"
	FieldDisableAdminChecking                 = "DisableAdminChecking"
	FieldStorageConfig                        = "StorageConfig"
	FieldDefaultStorageType                   = "DefaultStorageType"
	FieldGoogleAnalytics                      = "GoogleAnalytics"
	FieldFirebase                             = "Firebase"
	FieldApiHost                              = "ApiHost"
	FieldGames                                = "Games"
)

//ConfigModeFieldType is the type of field in RawConfigMode. Different
//UpdateConfig function factories only work on certain types of fields.
type ConfigModeFieldType int

const (
	FieldTypeInvalid ConfigModeFieldType = iota
	FieldTypeString
	FieldTypeStringSlice
	FieldTypeStringMap
	FieldTypeBool
	FieldTypeFirebase
	FieldTypeGameNode
)

//FieldTypes maps each ConfigModeField to its ConfigModeFieldType.
var FieldTypes = map[ConfigModeField]ConfigModeFieldType{
	FieldAllowedOrigins:       FieldTypeString,
	FieldDefaultPort:          FieldTypeString,
	FieldDefaultStaticPort:    FieldTypeString,
	FieldAdminUserIds:         FieldTypeStringSlice,
	FieldDisableAdminChecking: FieldTypeBool,
	FieldStorageConfig:        FieldTypeStringMap,
	FieldDefaultStorageType:   FieldTypeString,
	FieldGoogleAnalytics:      FieldTypeString,
	FieldFirebase:             FieldTypeFirebase,
	FieldApiHost:              FieldTypeString,
	FieldGames:                FieldTypeGameNode,
}

//ConfigModeCommon is the values that both ConfigMode and RawConfigMode share
//directly, factored out for convenience so they can be anonymously embedded
//in ConfigMdoe and RawConfigMode.
type ConfigModeCommon struct {
	AllowedOrigins    string   `json:"allowedOrigins,omitempty"`
	DefaultPort       string   `json:"defaultPort,omitempty"`
	DefaultStaticPort string   `json:"defaultStaticPort,omitempty"`
	AdminUserIds      []string `json:"adminUserIds,omitempty"`
	//This is a dangerous config. Only enable in Dev!
	DisableAdminChecking bool              `json:"disableAdminChecking,omitempty"`
	StorageConfig        map[string]string `json:"storageConfig,omitempty"`
	//The storage type that should be used if no storage type is provided via
	//command line options.
	DefaultStorageType string `json:"defaultStorageType,omitempty"`

	//The GA config string. Will be used to generate the client_config json
	//blob. Generally has a structure like "UA-321655-11"
	GoogleAnalytics string          `json:"googleAnalytics,omitempty"`
	Firebase        *FirebaseConfig `json:"firebase,omitempty"`
	//The host name the client should connect to in that mode. Something like
	//"http://localhost:8888"
	ApiHost string `json:"apiHost,omitempty"`
}
