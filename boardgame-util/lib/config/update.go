package config

import (
	"errors"
)

//ConfigUpdater is a function that takes a rawConfig and makes a modifcation
//in place on that object. It should return a non-nil error if it wasn't able
//to do the modification for some reason. These are one of the primary objects
//to config.Update(). This package defines a number of factories for these.
type ConfigUpdater func(r *RawConfigMode) error

//SetString returns a function to set the given rawconfig string property to
//the given value.
func SetString(field ConfigModeField, val string) ConfigUpdater {

	return func(r *RawConfigMode) error {
		switch field {
		case FieldAllowedOrigins:
			r.AllowedOrigins = val
		case FieldDefaultPort:
			r.DefaultPort = val
		case FieldDefaultStaticPort:
			r.DefaultStaticPort = val
		case FieldDefaultStorageType:
			r.DefaultStorageType = val
		case FieldGoogleAnalytics:
			r.GoogleAnalytics = val
		case FieldApiHost:
			r.ApiHost = val
		default:
			return errors.New(string(field) + " is not a valid string property")
		}

		return nil

	}
}

//DeleteString returns a function to unset the given rawcofngi string
//property.
func DeleteString(field ConfigModeField) ConfigUpdater {
	return SetString(field, "")
}
