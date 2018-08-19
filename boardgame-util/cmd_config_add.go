package main

import (
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	"strings"
)

type ConfigAdd struct {
	baseSubCommand

	Secret bool
	Dev    bool
	Prod   bool
}

func (c *ConfigAdd) Run(p writ.Path, positional []string) {

	base := c.Base().(*BoardgameUtil)

	cfg := base.GetConfig()

	mode := deriveMode(c.Dev, c.Prod)

	if len(positional) < 1 {
		errAndQuit("KEY must be provided")
	}

	field := config.FieldFromString(positional[0])

	fieldType := config.FieldTypes[field]

	var updater config.ConfigUpdater

	switch fieldType {
	case config.FieldTypeInvalid:
		errAndQuit(positional[0] + " is not a valid field")
	case config.FieldTypeStringSlice:
		if len(positional) != 2 {
			errAndQuit("KEY of type []string wants precisely one VAL")
		}
		updater = config.AddString(field, positional[1])
	case config.FieldTypeGameNode:
		errAndQuit("GAmes not yet supported")
	default:
		errAndQuit("Invalid field type for this command")
	}

	if err := cfg.Update(mode, c.Secret, updater); err != nil {
		errAndQuit("Couldn't update value: " + err.Error())
	}

	if err := cfg.Save(); err != nil {
		errAndQuit("Couldn't save updated config files: " + err.Error())
	}

}

func (c *ConfigAdd) Name() string {
	return "add"
}

func (c *ConfigAdd) Description() string {
	return "Adds the given value to the list for the given field for slice types"
}

func (c *ConfigAdd) WritOptions() []*writ.Option {
	return []*writ.Option{
		{
			Names:       []string{"s", "secret"},
			Flag:        true,
			Description: "If provided, will set the secret config instead of public config, creating it if necessary.",
			Decoder:     writ.NewFlagDecoder(&c.Secret),
		},
		{
			Names:       []string{"d", "dev"},
			Description: "If set, will write to dev options instead of base. No effect if prod is also passed",
			Flag:        true,
			Decoder:     writ.NewFlagDecoder(&c.Dev),
		},
		{
			Names:       []string{"p", "prod"},
			Description: "If set, will write to prod options instead of base. No effect if dev is also passed",
			Flag:        true,
			Decoder:     writ.NewFlagDecoder(&c.Prod),
		},
	}
}

func (c *ConfigAdd) Usage() string {
	return "[--secret] [--dev|--prod] KEY VAL"
}

func (c *ConfigAdd) HelpText() string {

	return c.Name() + " adds the given value to the list of the given field in the current config, for fields that are slice types. No op if that value is already in the list. KEY is not case sensitive.\n\n" +

		"If KEY is of type []string, simply adds the key to the given val if it doesn't exist. `config set` also has a similar effect for fields of this type. " +

		"Keys of this type are (" + strings.Join(keyNamesForConfigType(config.FieldTypeStringSlice), ",") + ")"

}
