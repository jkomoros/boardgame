package stub

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

//InteractiveOptions renders an interactve prompt at out, in to generate an
//Options from the user. If in or out are nil, StdIn or StdOut will be used
//implicitly.
func InteractiveOptions(in, out *os.File, gameName string) *Options {

	if in == nil {
		in = os.Stdin
	}

	if out == nil {
		out = os.Stdout
	}

	result := &Options{
		Name: gameName,
	}

	if displayName := getString(out, in, "What is the human-readable display name for this game? (e.g. 'Checkers', 'Tic Tac Toe')", result.Name); displayName != result.Name {
		result.DisplayName = displayName
	}

	numPlayersString := getString(out, in, "Range of valid players counts", "2-4")

	min, max, defaultNum, err := parseNumPlayers(numPlayersString)

	if err != nil {
		fmt.Println("That value is not valid: " + err.Error())
		return nil
	}

	result.MinNumPlayers = min
	result.MaxNumPlayers = max
	result.DefaultNumPlayers = defaultNum

	extras := getBool(out, in, "Generate useful extras like a test, CurrentPhase, and CurrentPlayer?", true)

	if !extras {
		result.SuppressExtras()
	}

	client := getBool(out, in, "Generate stub client renderers?", true)

	if !client {
		result.SuppressClient()
	}

	tutorial := getBool(out, in, "Generate extra tutorial content that demonstrate how to wire up more things?", false)

	if tutorial {
		result.EnableTutorials()
	}

	return result
}

func parseNumPlayers(in string) (min, max, defaultNum int, err error) {

	parts := strings.Split(in, "-")

	if len(parts) == 1 {
		return 0, 0, 0, errors.New("Invalid string, no '-'")
	}

	if len(parts) > 2 {
		return 0, 0, 0, errors.New("Too many '-'")
	}

	min, err = strconv.Atoi(strings.TrimSpace(parts[0]))

	if err != nil {
		return 0, 0, 0, errors.New("Min value is not an int: " + err.Error())
	}

	max, err = strconv.Atoi(strings.TrimSpace(parts[1]))

	if err != nil {
		return 0, 0, 0, errors.New("Max value is not an int: " + err.Error())
	}

	return min, max, min, nil

}

func getString(out, in *os.File, prompt, defaultValue string) string {

	if defaultValue != "" {
		prompt += " [" + defaultValue + "]"
	}

	prompt += ": "
	fmt.Fprint(out, prompt)
	var response string
	fmt.Fscanln(in, &response)

	response = strings.TrimSpace(response)

	if defaultValue != "" && response == "" {
		return defaultValue
	}

	return response
}

func getBool(out, in *os.File, message string, defaultVal bool) bool {
	if defaultVal {
		message += " [Y/n]"
	} else {
		message += " [y/N]"
	}

	response := getString(out, in, message, "")

	yesResponses := []string{"Yes", "Y", "yes", "y"}
	noResponses := []string{"No", "N", "no", "n"}

	if defaultVal {
		for _, responseToTest := range noResponses {
			if response == responseToTest {
				return false
			}
		}
		return true
	} else {
		for _, responseToTest := range yesResponses {
			if response == responseToTest {
				return true
			}
		}
	}
	return false
}
