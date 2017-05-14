package memory

import (
	"encoding/json"
	"github.com/jkomoros/boardgame"
	"log"
	"math/rand"
)

const DefaultMemoryLength = 6
const DefaultMemoryFuzziness = 0.03

const debugMode = false

type Agent struct{}

type agentCardInfo struct {
	Value string
	Index int
}

type agentState struct {
	LastCards []agentCardInfo
	//How many cards to remember
	MemoryLength int
	//How likely we are every time we see a card to forget the last one. Note
	//that we see cards often so this value should be pretty low.
	MemoryFuzziness float32
}

func (a *Agent) Name() string {
	return "ai"
}

func (a *Agent) DisplayName() string {
	return "Robby the Robot"
}

func (a *Agent) SetUpForGame(game *boardgame.Game, player boardgame.PlayerIndex) []byte {
	agent := &agentState{
		MemoryLength:    DefaultMemoryLength,
		MemoryFuzziness: DefaultMemoryFuzziness,
	}

	blob, err := json.MarshalIndent(agent, "", "\t")

	if err != nil {
		log.Println("Failued to marshal in set up for game", err)
		return nil
	}

	return blob
}

func (a *Agent) ProposeMove(game *boardgame.Game, player boardgame.PlayerIndex, aState []byte) (move boardgame.Move, newState []byte) {

	agent := &agentState{}

	if debugMode {
		log.Println(string(aState))
	}

	if err := json.Unmarshal(aState, agent); err != nil {
		log.Println("Failed to unmarshal agent state:", err)
		return nil, nil
	}

	state := game.CurrentState()

	gameState, _ := concreteStates(state)

	//Cull any cards taht are no longer valid
	doSave := agent.CullInvalidCards(gameState)

	//Take note of the position of any revealed cards
	for i, c := range gameState.RevealedCards.Components() {
		if c == nil {
			continue
		}
		card := c.Values.(*cardValue)
		if agent.CardSeen(card.Type, i) {
			if debugMode {
				log.Println("Card", card.Type, i, "is seen")
			}
			doSave = true
		}
	}

	if agent.PerhapsForgetCard() {
		if debugMode {
			log.Println("Forgetting a card")
		}
		doSave = true
	}

	if gameState.CurrentPlayer == player {

		//It's our turn!

		if gameState.RevealedCards.NumComponents() == 0 {
			//First card to reveal

			move = MoveRevealCardFactory(state)
			revealMove := move.(*MoveRevealCard)
			revealMove.CardIndex = agent.FirstCardToFlip(gameState)
			doSave = true
		}

		if gameState.RevealedCards.NumComponents() == 1 {

			//One more card to reveal

			move = MoveRevealCardFactory(state)
			revealMove := move.(*MoveRevealCard)
			revealMove.CardIndex = agent.SecondCardToFlip(gameState)
			doSave = true

		}

	}

	//Save back out agent state if something changed

	if doSave {
		//Marshal agent here
		var err error
		newState, err = json.MarshalIndent(agent, "", "\t")

		if err != nil {
			log.Println("Unable to marshal agent state:", err)
		}
	}

	return

}

func (a *agentState) PerhapsForgetCard() bool {
	if len(a.LastCards) < 1 {
		return false
	}
	if rand.Float32() < a.MemoryFuzziness {
		a.LastCards = a.LastCards[:len(a.LastCards)-1]
		return true
	}
	return false
}

//CullInvalidCards removes any remembered cards that no longer exist.
func (a *agentState) CullInvalidCards(gameState *gameState) bool {
	i := 0
	cardsCulled := false
	for i < len(a.LastCards) {
		card := a.LastCards[i]
		if c := gameState.HiddenCards.ComponentAt(card.Index); c != nil {
			//This card is still legit.
			i++
			continue
		}
		a.LastCards = append(a.LastCards[:i], a.LastCards[i+1:]...)
		cardsCulled = true
		//DON'T increment i; the next index is now i
	}
	return cardsCulled
}

//CardSeen is called when a card is visible. If will return true if that was
//new information, or false if not.
func (a *agentState) CardSeen(value string, index int) bool {

	//Is this card currently in the known set of cards?
	for _, card := range a.LastCards {
		if card.Value == value && card.Index == index {
			return false
		}
	}

	//Add it to the list.

	info := agentCardInfo{
		Value: value,
		Index: index,
	}

	a.LastCards = append([]agentCardInfo{info}, a.LastCards...)

	if len(a.LastCards) > a.MemoryLength {
		//Trim it if there are more cards than we can remember at once.
		a.LastCards = a.LastCards[:a.MemoryLength]
	}

	return true

}

//FlippedCard should be -1 if no cards are currently flipped.
func (a *agentState) FirstCardToFlip(gameState *gameState) int {
	//In our memory is there a pair?

	a.CullInvalidCards(gameState)

	seenValues := make(map[string]bool)
	valueToFlip := ""

	for _, card := range a.LastCards {
		if seenValues[card.Value] {
			//We saw two of the same cards, that's the one we want!
			valueToFlip = card.Value
			break
		}
		seenValues[card.Value] = true
	}

	if valueToFlip != "" {
		if debugMode {
			log.Println("Targeting card", valueToFlip, "for first card")
		}
		//Find the cards and return them.
		for _, card := range a.LastCards {
			if card.Value == valueToFlip {
				return card.Index
			}
		}
	}

	if debugMode {
		log.Println("Targeting random card for first card")
	}

	//Meh, we don't know which one to flip, flip any cards that haven't been
	//seen in memory and are not empty.
	return a.PickRandomCard(gameState)

}

func (a *agentState) PickRandomCard(gameState *gameState) int {
	for {
		index := rand.Intn(gameState.HiddenCards.Len())

		//Make sure that index actually is for a card that exists.
		if c := gameState.HiddenCards.ComponentAt(index); c == nil {
			continue
		}

		//Make sure that index isn't one we're already aware of.
		ok := true

		for _, card := range a.LastCards {
			if card.Index == index {
				ok = false
				break
			}
		}

		if !ok {
			continue
		}

		//OK, this sems like a good index.
		return index
	}

	return -1
}

func (a *agentState) SecondCardToFlip(gameState *gameState) int {
	a.CullInvalidCards(gameState)

	flippedCard := ""

	for _, c := range gameState.RevealedCards.Components() {
		if c == nil {
			continue
		}
		flippedCard = c.Values.(*cardValue).Type
	}

	if flippedCard == "" {
		log.Println("We were told to flip the second card but there wasn't already one flipped")
		return -1
	}

	//Is the flippedCard one where we know where the other one is?
	for _, card := range a.LastCards {
		if card.Value == flippedCard {
			if debugMode {
				log.Println("For second card targeting a card we know is there", card.Index)
			}
			return card.Index
		}
	}

	if debugMode {
		log.Println("For second card picking a random card")
	}

	//Otherwise, just pick a random card
	return a.PickRandomCard(gameState)
}
