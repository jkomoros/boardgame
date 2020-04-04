package behaviors

//Seat is a struct designed to be anonymously embedded into your PlayerState. It
//allows you to express that a given Player "slot" actually might be filled by a
//real player or not. When used in conjunction with moves.SeatPlayer, it allows
//your game logic to detect when new players are seated (when they join the
//game), as well as to control when that happens. Implements
//moves/interfaces.Seater
type Seat struct {
	SeatFilled bool
	SeatClosed bool
}

//SeatIsFilled returns true if a real player is sitting in this seat.
func (s *Seat) SeatIsFilled() bool {
	return s.SeatFilled
}

//SeatIsClosed returns true if the seat is closed--that is, no new people may
//sit in it, either because it is already filled, or because it has been
//affirmatively closed.
func (s *Seat) SeatIsClosed() bool {
	return s.SeatClosed
}

//SetSeatFilled sets that the seat is filled. Also sets SeatClosed to true.
func (s *Seat) SetSeatFilled() {
	s.SeatFilled = true
	s.SeatClosed = true
}

//SetSeatClosed sets that the seat is closed and should not be filled by any
//players, even if it is not filled. This tells the engine to not seat any more
//players here.
func (s *Seat) SetSeatClosed() {
	s.SeatClosed = true
}
