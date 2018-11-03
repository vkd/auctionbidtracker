package auctionbidtracker

// BidID - type of id of bid
type BidID int

// Bid - bid for item
type Bid struct {
	ID     BidID
	Value  int
	UserID UserID
}

// CheckWinner - check winner bid
func CheckWinner(old, new *Bid) bool {
	if old == nil {
		return true
	}
	return new.Value > old.Value
}

// UserID - type of id of user
type UserID int

// User - user
type User struct {
	ID UserID
}

// ItemID - typ of id of item
type ItemID int

// Item - item of auction
type Item struct {
	ID ItemID
}
