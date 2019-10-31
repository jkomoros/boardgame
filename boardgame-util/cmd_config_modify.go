package main

import (
	"fmt"

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

	Force bool
}

//If fieldType is one this responds to, should either teturn an updater or
//errAndQuit. Otherwise, OK to return nil to signal it's not a valid tyep.
//fieldType won't be TypeInvalid; that will already be screened out.
type updateFactory func(base *BoardgameUtil, field config.ModeField, fieldType config.ModeFieldType, positional []string) config.Updater

func (c *ConfigModify) RunWithUpdateFactory(p writ.Path, positional []string, factory updateFactory) {

	cfg := c.Base().GetConfig(c.Force)

	mode := deriveMode(c.Dev, c.Prod)

	if len(positional) < 1 {
		c.Base().errAndQuit("KEY must be provided")
	}

	field := config.FieldFromString(positional[0])

	fieldType := config.FieldTypes[field]

	if fieldType == config.FieldTypeInvalid {
		c.Base().errAndQuit(positional[0] + " is not a valid field")
	}

	if !c.ConfirmField(field) {
		c.Base().errAndQuit("Didn't confirm secret field set.")
	}

	updater := factory(c.Base(), field, fieldType, positional)

	if updater == nil {
		c.Base().errAndQuit("Invalid field type for this command")
	}

	if err := cfg.Update(mode, c.Secret, updater); err != nil {
		c.Base().errAndQuit("Couldn't update value: " + err.Error())
	}

	if err := cfg.Save(); err != nil {
		c.Base().errAndQuit("Couldn't save updated config files: " + err.Error())
	}

}

func (c *ConfigModify) ConfirmField(field config.ModeField) bool {

	//We only warn for secret fields on prod.
	if !c.Prod {

		if field == config.FieldOfflineDevMode {
			fmt.Println("WARNING: Typically you don't set offlinedevmode in your config, but pass --offline-dev-mode to 'boardgame-util serve' when you want it.")
		}

		return true
	}
	//Setting on secret is always fine.
	if c.Secret {
		return true
	}

	sensitiveFields := map[config.ModeField]bool{
		config.FieldAdminUserIds: true,
		config.FieldStorage:      true,
	}

	//Only show confirm for sensitive fields.
	if !sensitiveFields[field] {
		return true
	}

	return baseConfirm("You have proposed setting a field that is typically secret on prod without passing `--secret`, which means the configuration would be added to the public config that might be checked into source control.")

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
		{
			Names:       []string{"f", "force"},
			Description: "If provided, will allow modifying the given property even if the underlying config does not yet exist.",
			Flag:        true,
			Decoder:     writ.NewFlagDecoder(&c.Force),
		},
	}
}

func (c *ConfigModify) Usage() string {
	return "[--secret] [--dev|--prod] KEY [SUB-KEY] VAL"
}
