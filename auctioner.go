package auctionbidtracker

// Auctioner - interface for auction
type Auctioner interface {
	MakeBid(itemID ItemID, userID UserID, bid *Bid) error
	GetWinningBid(itemID ItemID) (*Bid, error)
	GetAllBids(itemID ItemID) ([]*Bid, error)
	GetItemsByUser(userID UserID) ([]*Item, error)

	AddItem(item *Item) (*Item, error)
}
