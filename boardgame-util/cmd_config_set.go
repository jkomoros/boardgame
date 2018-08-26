package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	"strings"
)

type ConfigSet struct {
	ConfigModify
}

func strToBool(in string) (bool, error) {

	in = strings.ToLower(in)

	if in == "0" || in == "false" {
		return false, nil
	}

	if in == "1" || in == "true" {
		return true, nil
	}

	return false, errors.New("Invalid bool string: " + in)

}

func deriveMode(dev, prod bool) config.ConfigModeType {

	//If one of dev or prod is set, return that
	if dev || prod {
		if dev {
			return config.TypeDev
		}
		if prod {
			return config.TypeProd
		}
	}

	//Default to base
	return config.TypeBase
}

func configSetFactory(field config.ConfigModeField, fieldType config.ConfigModeFieldType, positional []string) config.ConfigUpdater {
	switch fieldType {
	case config.FieldTypeString:
		if len(positional) != 2 {
			errAndQuit("KEY of type string wants precisely one VAL")
		}
		return config.SetString(field, positional[1])
	case config.FieldTypeStringSlice:
		if len(positional) != 2 {
			errAndQuit("KEY of type []string wants precisely one VAL")
		}
		return config.AddString(field, positional[1])
	case config.FieldTypeStringMap:
		if len(positional) != 3 {
			errAndQuit("KEY of type map[string]string wants KEY SUB-KEY VAL")
		}
		return config.SetStringKey(field, positional[1], positional[2])
	case config.FieldTypeBool:
		if len(positional) != 2 {
			errAndQuit("KEY of type bool wants one VAL")
		}
		b, err := strToBool(positional[1])
		if err != nil {
			errAndQuit(err.Error())
		}
		return config.SetBool(field, b)
	case config.FieldTypeFirebase:
		if len(positional) != 3 {
			errAndQuit("KEY of type firebase wants KEY SUB-KEY VAL")
		}

		firebaseKey := config.FirebaseKeyFromString(positional[1])

		if firebaseKey == config.FirebaseInvalid {
			errAndQuit(positional[1] + " is not a valid firebase key")
		}

		return config.SetFirebaseKey(firebaseKey, positional[2])
	case config.FieldTypeGameNode:
		if len(positional) != 2 {
			errAndQuit("games node wants precisely one VAL")
		}
		return config.AddGame(positional[1])
	}
	return nil
}

func (c *ConfigSet) Run(p writ.Path, positional []string) {
	c.ConfigModify.RunWithUpdateFactory(p, positional, configSetFactory)
}

func (c *ConfigSet) Name() string {
	return "set"
}

func (c *ConfigSet) Description() string {
	return "Sets the given field to the given value"
}

func keyNamesForConfigType(typ config.ConfigModeFieldType) []string {

	var result []string

	for key, val := range config.FieldTypes {
		if val == typ {
			result = append(result, "'"+string(key)+"'")
		}
	}

	return result

}

func (c *ConfigSet) HelpText() string {

	var firebaseKeys []string

	for key := range config.FirebaseKeys {
		firebaseKeys = append(firebaseKeys, "'"+string(key)+"'")
	}

	return c.Name() + " sets the given field to the given value in the current config. KEY is not case sensitive.\n\n" +

		"If KEY is of type string, simply sets the key to the given val. " +

		"Keys of this type are (" + strings.Join(keyNamesForConfigType(config.FieldTypeString), ",") + ")\n\n" +

		"If KEY is of type []string, simply adds the key to the given val if it doesn't exist. A simple convenience instead of `config add`. " +

		"Keys of this type are (" + strings.Join(keyNamesForConfigType(config.FieldTypeStringSlice), ",") + "). `Games` is also of this type.\n\n" +

		"If KEY is of type bool, val must be either '0', '1', 'true', 'false'. " +

		"Keys of this type are (" + strings.Join(keyNamesForConfigType(config.FieldTypeBool), ",") + ")\n\n" +

		"If KEY is of type map[key]val then SUB-KEY must also be provided. A value of '' will delete the key. " +

		"Keys of this type are (" + strings.Join(keyNamesForConfigType(config.FieldTypeStringMap), ",") + "). " +

		"'Firebase' is also of this type, but only allows the following sub-keys: (" +

		strings.Join(firebaseKeys, ",") + "), and keys set to value of '' will not be deleted.\n\n" +

		"See help for the parent command for more information about configuration in general."
}
