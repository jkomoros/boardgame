package main

import (
	"github.com/bobziuchkovski/writ"
)

//SubcommandObject is a literal struct that implements a subcommand
type SubcommandObject interface {
	//The name the command is registered as
	Name() string
	//The description of the command to register
	Description() string

	//HelpText is the prose decription of what this command does.
	HelpText() string

	//The aliases to register for
	Aliases() []string
	//The rest of the usage string, which will be appened to "NAME "
	Usage() string
	//SubcommandObjects should return the list of sub comamnds, or nil if a
	//terminal command.
	SubcommandObjects() []SubcommandObject
	//The command to actually run
	Run(p writ.Path, positional []string)

	//Config returns a writ.Command object. Should return the same object on
	//repeated calls.
	WritCommand() *writ.Command
	WritOptions() []*writ.Option
	//WritparentOptions returns a series of frames back up to the root command
	//with options. They aren't regeistered, per se, but are used to generate
	//help for all options.
	WritparentOptions() []*parentOptions
	//WritHelp should return a Help config object for this command
	WritHelp() writ.Help

	TopLevelStruct() SubcommandObject
	SetTopLevelStruct(top SubcommandObject)

	Base() *boardgameUtil
	SetBase(base *boardgameUtil)

	Parent() SubcommandObject
	//SetParent will be called with the command's parent object.
	SetParent(parent SubcommandObject)
}

type parentOptions struct {
	Name    string
	Options []*writ.Option
	Cmd     *writ.Command
}

type baseSubCommand struct {
	parent         SubcommandObject
	topLevelStruct SubcommandObject
	writCommand    *writ.Command
	base           *boardgameUtil
}

func (b *baseSubCommand) Base() *boardgameUtil {
	return b.base
}

func (b *baseSubCommand) SetBase(base *boardgameUtil) {
	b.base = base
}

func (b *baseSubCommand) optionGroupForObject(name string, cmd *writ.Command, options []*writ.Option) *writ.OptionGroup {

	if len(options) == 0 {
		return nil
	}

	optionNames := make([]string, len(options))
	for i, opt := range options {
		optionNames[i] = opt.Names[0]
	}
	group := cmd.GroupOptions(optionNames...)
	if name == "" {
		group.Header = "Options:"
	} else {
		group.Header = "Options for " + name + ":"
	}

	return &group

}

func (b *baseSubCommand) WritHelp() writ.Help {

	if b.WritCommand() == nil {
		return writ.Help{}
	}

	obj := b.TopLevelStruct()

	//TODO: pop this in as well
	var result writ.Help

	result.Header = obj.HelpText()

	baseSubCommands := obj.SubcommandObjects()

	if len(baseSubCommands) > 0 {

		subCmdNames := make([]string, len(baseSubCommands))
		for i, obj := range baseSubCommands {
			subCmdNames[i] = obj.Name()
		}

		group := b.WritCommand().GroupCommands(subCmdNames...)
		group.Header = "Subcommands:"
		result.CommandGroups = append(result.CommandGroups, group)

	}

	group := b.optionGroupForObject("", obj.WritCommand(), obj.WritOptions())
	if group != nil {
		result.OptionGroups = append(result.OptionGroups, *group)
	}

	for _, parentOptions := range obj.WritparentOptions() {
		group := b.optionGroupForObject(parentOptions.Name, parentOptions.Cmd, parentOptions.Options)
		if group != nil {
			result.OptionGroups = append(result.OptionGroups, *group)
		}
	}

	result.Usage = "Usage: " + fullName(obj) + " " + obj.Usage()

	return result
}

func (b *baseSubCommand) WritCommand() *writ.Command {
	if b.writCommand != nil {
		return b.writCommand
	}

	obj := b.TopLevelStruct()

	subCommands := obj.SubcommandObjects()
	subConfigs := make([]*writ.Command, len(subCommands))
	for i, command := range subCommands {
		subConfigs[i] = command.WritCommand()
	}

	config := &writ.Command{
		Name:        obj.Name(),
		Description: obj.Description(),
		Aliases:     obj.Aliases(),
		Subcommands: subConfigs,
		Options:     obj.WritOptions(),
	}

	b.writCommand = config

	config.Help = obj.WritHelp()

	return config
}

func (b *baseSubCommand) WritparentOptions() []*parentOptions {
	var result []*parentOptions

	obj := b.TopLevelStruct()

	obj = obj.Parent()

	for obj != nil {

		parentOptions := &parentOptions{
			Name:    obj.Name(),
			Options: obj.WritOptions(),
			Cmd:     obj.WritCommand(),
		}

		result = append(result, parentOptions)

		obj = obj.Parent()
	}

	return result
}

func (b *baseSubCommand) TopLevelStruct() SubcommandObject {
	return b.topLevelStruct
}

func (b *baseSubCommand) SetTopLevelStruct(top SubcommandObject) {
	b.topLevelStruct = top
}

func (b *baseSubCommand) SetParent(parent SubcommandObject) {
	b.parent = parent
}

func (b *baseSubCommand) Parent() SubcommandObject {
	return b.parent
}

func (b *baseSubCommand) Aliases() []string {
	return nil
}

func (b *baseSubCommand) Description() string {
	return ""
}

func (b *baseSubCommand) Usage() string {
	return ""
}

//HelpText defaults to description
func (b *baseSubCommand) HelpText() string {
	return b.TopLevelStruct().Description()
}

func (b *baseSubCommand) SubcommandObjects() []SubcommandObject {
	return nil
}

func (b *baseSubCommand) WritOptions() []*writ.Option {
	return nil
}

func setupParents(cmd SubcommandObject, parent SubcommandObject, base *boardgameUtil) {

	cmd.SetParent(parent)
	cmd.SetTopLevelStruct(cmd)
	cmd.SetBase(base)

	if parent == nil {
		base = cmd.(*boardgameUtil)
	}

	for _, subCmd := range cmd.SubcommandObjects() {
		setupParents(subCmd, cmd, base)
	}

}

func fullName(cmd SubcommandObject) string {
	if cmd.Parent() == nil {
		return cmd.Name()
	}
	return fullName(cmd.Parent()) + " " + cmd.Name()
}

func strMatchesObject(str string, s SubcommandObject) bool {
	if s.Name() == str {
		return true
	}

	for _, alias := range s.Aliases() {
		if alias == str {
			return true
		}
	}

	return false
}

//selectSubcommandObject takes a subcommand object and a path. It verifes the
//first item is us, then identifies the nexct object to recurse into based on
//Names of SubcommandObjects.
func selectSubcommandObject(s SubcommandObject, p []string) SubcommandObject {

	if !strMatchesObject(p[0], s) {
		return nil
	}

	if len(p) < 2 {
		return s
	}

	nextCommand := p[1]

	for _, obj := range s.SubcommandObjects() {
		//We don't need to check alises, because the main library already did
		//the command/object matching
		if strMatchesObject(nextCommand, obj) {
			return selectSubcommandObject(obj, p[1:])
		}
	}

	return nil
}
