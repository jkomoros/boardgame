package main

import (
	"fmt"
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	"strings"
)

type ConfigSet struct {
	baseSubCommand

	Secret bool
	Mode   string
}

func (c *ConfigSet) Run(p writ.Path, positional []string) {

	//base := c.Base().(*BoardgameUtil)

	//cfg := base.GetConfig()

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

	fmt.Println("This command is not yet fully functional")
	fmt.Println("Would have set", mode, c.Secret)

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

//TODO: usage for positional

func (c *ConfigSet) HelpText() string {
	return c.Name() + ` sets the given field to the given value in the current config.`
}
