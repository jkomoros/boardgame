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

	type RepeatTokenComponent struct {
		Repeat int
		Component *TokenComponent
	}

	type myChest struct {
		CardDeck []*CardComponent
		TokenDeck []*RepeatTokenComponent
	}

Where *CardComponent and *TokenComponent both implement
boardgame.ComponentValues. Note that if the value is a struct with two fields,
Repeat and Component, Repeat number of the component will be added to the
deck.

At this point, the RepeatComponents and the Components themselves must be
pointers.

*/
func FromConfig(blob []byte, container interface{}) (*boardgame.ComponentChest, error) {

	if err := json.Unmarshal(blob, container); err != nil {
		return nil, errors.New("The provided blob could not be parsed: " + err.Error())
	}

	s := reflect.ValueOf(container).Elem()
	typeOfContainer := s.Type()

	chest := boardgame.NewComponentChest()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		if f.Kind() != reflect.Slice {
			//We ignore anything that's not a slice.
			continue
		}

		deck := boardgame.NewDeck()

		for j := 0; j < f.Len(); j++ {

			val := f.Index(j)
			typeOfVal := val.Elem().Type()

			simpleComponent := true

			//Check to see if this is a repeat object
			if val.Elem().NumField() == 2 {
				nameZero := typeOfVal.Field(0).Name
				nameOne := typeOfVal.Field(1).Name
				if (nameOne == "Repeat" && nameZero == "Component") || (nameZero == "Repeat" && nameOne == "Component") {

					//It's a repeat component!
					deck.AddComponentMulti(val.Elem().FieldByName("Component").Interface().(boardgame.SubState), val.Elem().FieldByName("Repeat").Interface().(int))

					//Signal that we already added it.
					simpleComponent = false
				}
			}

			if simpleComponent {
				deck.AddComponent(
					//TODO: verify this works so we don't panic
					val.Interface().(boardgame.SubState),
				)
			}
		}

		chest.AddDeck(typeOfContainer.Field(i).Name, deck)

	}

	chest.Finish()

	return chest, nil

}
