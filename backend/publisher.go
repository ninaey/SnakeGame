package main

import (
	"fmt"
)

// Represents the "Skin" and "Lives" box from your diagram
type GameState struct {
	Skin         string
	Lives        int
	CurrentScore int
}

// PA: The Publisher App component
type Publisher struct {
	Data GameState
	Out  chan string // The "Arrow" pointing to G
}

// NewPublisher initializes the PA with your starting values
func NewPublisher(outChannel chan string) *Publisher {
	return &Publisher{
		Data: GameState{
			Skin:  "Default",
			Lives: 2, // Your "2rm"
		},
		Out: outChannel,
	}
}

// SendUpdate publishes the current state to the G box
func (p *Publisher) SendUpdate(event string) {
	message := fmt.Sprintf("[%s] Skin: %s | Lives: %d | Score: %d",
		event, p.Data.Skin, p.Data.Lives, p.Data.CurrentScore)

	// Non-blocking send ensures the snake doesn't lag
	select {
	case p.Out <- message:
	default:
		// Drop message if G is overloaded to maintain game speed
	}
}
