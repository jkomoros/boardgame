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

	TopLevelStruct() SubcommandObject
	SetTopLevelStruct(top SubcommandObject)

	Parent() SubcommandObject
	//SetParent will be called with the command's parent object.
	SetParent(parent SubcommandObject)
}

type baseSubCommand struct {
	parent         SubcommandObject
	topLevelStruct SubcommandObject
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

func setupParents(cmd SubcommandObject, parent SubcommandObject) {

	cmd.SetParent(parent)
	cmd.SetTopLevelStruct(cmd)

	for _, subCmd := range cmd.SubcommandObjects() {
		setupParents(subCmd, cmd)
	}

}

func FullName(cmd SubcommandObject) string {
	if cmd.Parent() == nil {
		return cmd.Name()
	}
	return FullName(cmd.Parent()) + " " + cmd.Name()
}
