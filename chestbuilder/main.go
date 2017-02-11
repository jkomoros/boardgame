/*

chestbuilder is a simple package that makes it easy to generate a component
chest from a JSON blob.

*/
package chestbuilder

import (
	"encoding/json"
	"errors"
	"github.com/jkomoros/boardgame"
	"reflect"
)

/*
FromConfig takes in a JSON blob in bytes, loads it up into Container, and
then creates and returns a Chest by walking that container.

You container should look like:
	type myChest struct {
		CardDeck []*CardComponent
		ResourceTokenDeck []*ResourceTokenDeck
	}

Where *CardComponent and *ResourceTokenComponent both implement
boardgame.ComponentValues

*/
func FromConfig(blob []byte, container interface{}) (*boardgame.ComponentChest, error) {

	if err := json.Unmarshal(blob, container); err != nil {
		return nil, errors.New("The provided blob could not be parsed: " + err.Error())
	}

	s := reflect.ValueOf(container).Elem()
	typeOfContainer := s.Type()

	chest := boardgame.NewComponentChest("WHATEVER")

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		if f.Kind() != reflect.Slice {
			//We ignore anything that's not a slice.
			continue
		}

		deck := &boardgame.Deck{}

		for j := 0; j < f.Len(); j++ {
			deck.AddComponent(
				//TODO: verify this works so we don't panic
				f.Index(j).Interface().(boardgame.PropertyReader),
			)
		}

		chest.AddDeck(typeOfContainer.Field(i).Name, deck)

	}

	chest.Finish()

	return chest, nil

}
