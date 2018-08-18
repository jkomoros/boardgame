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
//the given value. field must be of FieldTypeString.
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

//DeleteString returns a function to unset the given config string
//propert, as long as field is of type FieldTypeString.
func DeleteString(field ConfigModeField) ConfigUpdater {
	return SetString(field, "")
}

//AddString adds the given string, if it doesn't exist, to the []string type
//ConfigModeField. Field must be of FieldTypeStringSlice.
func AddString(field ConfigModeField, val string) ConfigUpdater {

	return func(r *RawConfigMode) error {
		if field != FieldAdminUserIds {
			return errors.New(string(field) + " is not a []string field")
		}
		//Make sure the value isn't already set
		for _, rec := range r.AdminUserIds {
			if rec == val {
				//Already exists, we're done
				return nil
			}
		}

		r.AdminUserIds = append(r.AdminUserIds, val)
		return nil
	}

}

//RemoveString removes the given string, if it exists, from the []string type
//ConfigModeField. If it was the last item to remove, sets that field to nil.
//Field must be of FieldTypeStringSlice.
func RemoveString(field ConfigModeField, val string) ConfigUpdater {

	return func(r *RawConfigMode) error {
		if field != FieldAdminUserIds {
			return errors.New(string(field) + " is not a []string field")
		}
		var newList []string
		//Make sure the value isn't already set
		for _, rec := range r.AdminUserIds {
			if rec == val {
				//Don't copy this one over!
				continue
			}
			newList = append(newList, rec)
		}

		r.AdminUserIds = newList
		return nil
	}

}

//SetStringKey adds the given key and val to the map[string]string field
//denoted by field. If that key is already set, it updates it to the new
//value. If the map is nil, creates one. Field must be of FieldTypeStringMap.
func SetStringKey(field ConfigModeField, key, val string) ConfigUpdater {

	return func(r *RawConfigMode) error {
		if field != FieldStorageConfig {
			return errors.New(string(field) + " is not a map[string]string")
		}
		if r.StorageConfig == nil {
			r.StorageConfig = make(map[string]string)
		}
		r.StorageConfig[key] = val
		return nil
	}

}

//DeleteStringKey deletes the given key from the map[string]string field
//denoted by field. Field must be of FieldTypeStringMap.
func DeleteStringKey(field ConfigModeField, key string) ConfigUpdater {

	return func(r *RawConfigMode) error {
		if field != FieldStorageConfig {
			return errors.New(string(field) + " is not a map[string]string")
		}
		//This shouldn't happen
		if r.StorageConfig == nil {
			return nil
		}
		delete(r.StorageConfig, key)
		//Don't nil out the storage map if no keys are left; Storage should
		//always be non-nil.
		return nil
	}

}
