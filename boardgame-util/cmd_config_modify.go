package main

import (
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
)

//ConfigModify isn't used directly but is a super-class for all Config
//commands that modify the config.
type ConfigModify struct {
	baseSubCommand

	Secret bool
	Dev    bool
	Prod   bool
}

func (c *ConfigModify) Run(p writ.Path, positional []string) {

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

func (c *ConfigModify) WritOptions() []*writ.Option {
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

func (c *ConfigModify) Usage() string {
	return "[--secret] [--dev|--prod] KEY [SUB-KEY] VAL"
}
