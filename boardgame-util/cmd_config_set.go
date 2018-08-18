package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	"strings"
)

type ConfigSet struct {
	baseSubCommand

	Secret bool
	Mode   string
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

func (c *ConfigSet) Run(p writ.Path, positional []string) {

	base := c.Base().(*BoardgameUtil)

	cfg := base.GetConfig()

	mode := config.TypeBase

	c.Mode = strings.ToLower(c.Mode)

	switch c.Mode {
	case "base":
		//Pass; mode is already base
	case "dev":
		mode = config.TypeDev
	case "prod":
		mode = config.TypeProd
	default:
		errAndQuit(c.Mode + " is not a valid mode")
	}

	if len(positional) < 1 {
		errAndQuit("KEY must be provided")
	}

	field := config.FieldFromString(positional[0])

	fieldType := config.FieldTypes[field]

	var updater config.ConfigUpdater

	switch fieldType {
	case config.FieldTypeInvalid:
		errAndQuit(positional[0] + " is not a valid field")
	case config.FieldTypeString:
		if len(positional) != 2 {
			errAndQuit("KEY of type string wants precisely one VAL")
		}
		updater = config.SetString(field, positional[1])
	case config.FieldTypeStringSlice:
		if len(positional) != 2 {
			errAndQuit("KEY of type []string wants precisely one VAL")
		}
		updater = config.AddString(field, positional[1])
	case config.FieldTypeStringMap:
		if len(positional) != 3 {
			errAndQuit("KEY of type map[string]string wants KEY SUB-KEY VAL")
		}
		updater = config.SetStringKey(field, positional[1], positional[2])
	case config.FieldTypeBool:
		if len(positional) != 2 {
			errAndQuit("KEY of type bool wants one VAL")
		}
		b, err := strToBool(positional[1])
		if err != nil {
			errAndQuit(err.Error())
		}
		updater = config.SetBool(field, b)
	case config.FieldTypeFirebase:
		if len(positional) != 3 {
			errAndQuit("KEY of type firebase wants KEY SUB-KEY VAL")
		}

		firebaseKey := config.FirebaseKeyFromString(positional[1])

		if firebaseKey == config.FirebaseInvalid {
			errAndQuit(positional[1] + " is not a valid firebase key")
		}

		updater = config.SetFirebaseKey(firebaseKey, positional[2])
	case config.FieldTypeGameNode:
		errAndQuit("GAmes not yet supported")
	}

	if err := cfg.Update(mode, c.Secret, updater); err != nil {
		errAndQuit("Couldn't update value: " + err.Error())
	}

	if err := cfg.Save(); err != nil {
		errAndQuit("Couldn't save updated config files: " + err.Error())
	}

}

func (c *ConfigSet) Name() string {
	return "set"
}

func (c *ConfigSet) Description() string {
	return "Sets the given field to the given value"
}

func (c *ConfigSet) WritOptions() []*writ.Option {
	return []*writ.Option{
		{
			Names:       []string{"s", "secret"},
			Flag:        true,
			Description: "If provided, will set the secret config instead of public config, creating it if necessary.",
			Decoder:     writ.NewFlagDecoder(&c.Secret),
		},
		{
			Names:       []string{"m", "mode"},
			Description: "The mode type to operate on. One of {base, dev, prod}. Defaults to base.",
			Decoder: writ.NewDefaulter(
				writ.NewOptionDecoder(&c.Mode),
				"base",
			),
		},
	}
}

func (c *ConfigSet) Usage() string {
	return "KEY [SUB-KEY] VAL"
}

func (c *ConfigSet) HelpText() string {
	return c.Name() + ` sets the given field to the given value in the current config.

If KEY is of type string, simply sets the key to the given val.

If KEY is of type []string, simply adds the key to the given val if it doesn't exist.

If KEY is of type bool, val must be either "0", "1", "true", "false"

If KEY is of type map[key]val then SUB-KEY must also be provided.`
}
