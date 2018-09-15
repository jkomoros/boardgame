package boardgame

import (
	"errors"
	"strings"
)

//Variant is just a map of keys to values that are passed to your game so it
//can configure different alternate rulesets, for example using a Short
//variant that uses fewer cards and should play faster, or using a different
//deck of cards than normal. The variant configuration will be considered
//legal if it passes Delegate.Variants().LegalVariant(), and will be passed to
//Delegate.BeginSetup so that you can set up your game in whatever way makes
//sense for a given Variant. Your Delegate defines what valid keys and values
//are, and how they should be displayed to end-users, with its return value
//for Variants().
type Variant map[string]string

/*
VariantConfig defines the legal keys, and their legal values, that a variant
may have in this game. You return one of these from your delegate's
Variants() method. Your VariantConfig also defines the display name and
description for each key and each value that will be displayed to the end-
user.

The behavior of Initialize() (and the methods on sub-objects it calls) allows
you to omit certain configuration and have them automatically set.


	func (g *gameDelegate) Variants() VariantConfig {

		//As long as you access this via gameManager.Variants() instead of
		//directly from the delegate, Initialize will have already been called
		//for us.

		return VariantConfig{
			"color": {

				//You can skip setting the VariantDiplayInfo.Name,
				//.DisplayName here because initialize (which we call at the
				//end of this method) will automatically use the name of the
				//entry in the map, and then the displayname will be set to a
				//reasonable title-casing.

				Values: map[string]*VariantDisplayInfo{
					"red": {
						//Name can be omitted because Initialize() will
						//automatically set it bassed on this value's name in
						//the map.

						//Because DisplayName has been set expclitily it will
						//not be overriden in Initialize.
						DisplayName: "Very Red",
						Description: "The color red",
					},
					//You can leave the value empty, which will automatically
					//create a new value during Initalize with the Name coming
					//from the map, and DisplayName set automatically.
					"blue": nil,
				},

				//By setting this, any new Variant created from our NewVariant
				//will always have the "color" key to either the value
				//provided, or "blue".
				Default: "blue",
			},
		}

	}
*/
type VariantConfig map[string]*VariantKey

//VariantKey represents a specific key in your variants that has a particular
//meaning. For example, "color".
type VariantKey struct {
	//VariantKey has a DisplayInfo embedded in it the defines the display name
	//and description for this configuration key.
	VariantDisplayInfo
	//The name of the value, in Values, that is default if none provided. Must
	//exist in the Values map or Valid() will error.
	Default string
	//The specific values this key may take, along with their display
	//information. For example, "blue", "red".
	Values map[string]*VariantDisplayInfo
}

//VariantDisplayInfo is information about a given value and how to display it to end-
//users. It is used as part of VariantKey both to describe the Key itself as
//well as to give information about the values within the key.
type VariantDisplayInfo struct {
	DisplayName string
	Description string
	//Name can often be skipped because it is often set implicitly during
	//initialization of the containing object.
	Name string
}

//Valid returns an error if there is any misconfiguration in this
//VariantConfig. In particular, it verifies that the implied name for each key
//matches its explicit Name property, and the same for values. It also
//verifies that if there's a default it denotes a valid value. Effectively
//this checks if Initialize() has been called or not. NewGameManager will
//check that the config returned from Variants() passes this test and will
//fail to create if not, which is mainly to help you make sure you remember to
//call Initalize.
func (v VariantConfig) Valid() error {
	if v == nil {
		return nil
	}
	for name, key := range v {
		if name != key.Name {
			return errors.New("Key " + name + " does not have its name set the same: " + key.Name)
		}

		if len(key.Values) == 0 {
			return errors.New("Key " + name + " does not define any values.")
		}

		for valName, val := range key.Values {
			if val == nil {
				return errors.New("Key " + name + " value " + valName + " is set to nil")
			}
			if valName != val.Name {
				return errors.New("Key " + name + " value " + valName + " does not have its name set the same: " + val.Name)
			}
		}

		if key.Default != "" && key.Values[key.Default] == nil {
			return errors.New("Key " + name + " has a default of " + key.Default + " but that is not valid value")
		}
	}

	return nil
}

//Initalize calls initalize on each Key in config. Generally the game engine
//calls this and you don't have to call it in your delegate's Variants()
//method.
func (v VariantConfig) Initalize() {
	for key, val := range v {
		val.Initialize(key)
	}
}

//Initialize is given the name of this key within its parent's map. The
//provided name will override whatever Name was already set and also sets the
//display name (see VariantDisplayInfo.Initialize). Also calls Initialize() on
//all values.
func (v *VariantKey) Initialize(nameInParent string) {
	for key, val := range v.Values {
		if val == nil {
			val = &VariantDisplayInfo{}
			v.Values[key] = val
		}
		val.Initialize(key)
	}
	v.VariantDisplayInfo.Initialize(nameInParent)
}

//Initialize sets the name to the given name. It also sets the display name
//automatically if one wasn't provided by replacing "_" and "-" with spaces
//and title casing name.
func (d *VariantDisplayInfo) Initialize(nameInParent string) {
	d.Name = nameInParent

	if d.DisplayName != "" {
		return
	}

	displayName := d.Name

	displayName = strings.Replace(displayName, "-", " ", -1)
	displayName = strings.Replace(displayName, "_", " ", -1)

	//Reduce runs of whitespace to just a single space
	displayName = strings.Join(strings.Fields(displayName), " ")

	d.DisplayName = strings.Title(displayName)

}

//NewVariant returns a new varient with the given values set. Any extra keys
//that are not in VariantConfig will lead to an error, as well as any values
//that are illegal for their key. Any missing key/value pairs will be set to
//their default, if the key has a default. Typically you don't call this
//directly, but it is called for you implicitly within NewGame.
func (v VariantConfig) NewVariant(variantValues map[string]string) (Variant, error) {

	if len(variantValues) == 0 {
		return nil, nil
	}

	result := make(Variant, len(variantValues))

	for key, val := range variantValues {
		result[key] = val
	}

	for key, values := range v {
		if result[key] != "" {
			continue
		}
		if values.Default == "" {
			continue
		}
		result[key] = values.Default
	}

	if err := v.LegalVariant(result); err != nil {
		return nil, err
	}

	return result, nil

}

//LegalVariant returns whether the given variant has keys and values that are
//enumerated and legal in this config. In paticular, ensures that every key in
//variant is defined in this config, and the value for each key is one of the
//legal values according to the config. Nil configs are OK.
func (v VariantConfig) LegalVariant(variant Variant) error {

	if v == nil {
		if len(variant) == 0 {
			return nil
		}
		return errors.New("Variant defined values, but the VariantConfig in use didn't define any")
	}

	for key, val := range variant {
		configKey := v[key]
		if configKey == nil {
			return errors.New("configuration had a property called " + key + " that isn't expected")
		}
		configValue := configKey.Values[val]
		if configValue == nil {
			return errors.New("configuration's " + configKey.DisplayName + " property had a value that wasn't allowed: " + val)
		}
	}

	return nil

}
