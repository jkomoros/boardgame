package config

import (
	"errors"
	"strings"
)

//ConfigUpdater is a function that takes a rawConfig and makes a modifcation
//in place on that object. It should return a non-nil error if it wasn't able
//to do the modification for some reason. These are one of the primary objects
//to config.Update(). This package defines a number of factories for these.
type ConfigUpdater func(r *RawConfigMode) error

//SetString returns a function to set the given rawconfig string property to
//the given value.
func SetString(propName, val string) ConfigUpdater {

	propName = strings.ToLower(propName)
	propName = strings.TrimSpace(propName)

	return func(r *RawConfigMode) error {
		switch propName {
		case "allowedorigins":
			r.AllowedOrigins = val
		case "defaultport":
			r.DefaultPort = val
		case "defaultstaticport":
			r.DefaultStaticPort = val
		case "defaultstoragetype":
			r.DefaultStorageType = val
		case "googleanalytics":
			r.GoogleAnalytics = val
		case "apihost":
			r.ApiHost = val
		default:
			return errors.New(propName + " is not a valid string property")
		}

		return nil

	}
}
