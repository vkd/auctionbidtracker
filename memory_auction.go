package auctionbidtracker

import (
	"errors"
	"sync"
)

// MemoryAuction - auction with storage in memory
//
// This implementation of auction has one global mutex
// and has O(n) complexity on get-items-by-user method
type MemoryAuction struct {
	mx sync.Mutex

	bids map[ItemID]*itemBids
}

type itemBids struct {
	item      *Item
	winnerBid *Bid
	bids      []*Bid
}

var _ Auctioner = (*MemoryAuction)(nil)

// NewMemoryAuction - create new memoty auction
func NewMemoryAuction() *MemoryAuction {
	return &MemoryAuction{
		bids: make(map[ItemID]*itemBids),
	}
}

// MakeBid - make new bid for item
func (m *MemoryAuction) MakeBid(itemID ItemID, userID UserID, bid *Bid) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	bid.UserID = userID

	bs, ok := m.bids[itemID]
	if !ok {
		return errors.New("item not found")
	}
	bs.bids = append(bs.bids, bid)
	if CheckWinner(bs.winnerBid, bid) {
		bs.winnerBid = bid
	}

	return nil
}

// GetWinningBid - get winning bid of item
func (m *MemoryAuction) GetWinningBid(itemID ItemID) (*Bid, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	if bid, ok := m.bids[itemID]; ok {
		return bid.winnerBid, nil
	}
	return nil, nil
}

// GetAllBids - return list of bids
func (m *MemoryAuction) GetAllBids(itemID ItemID) ([]*Bid, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	if bid, ok := m.bids[itemID]; ok {
		return bid.bids, nil
	}
	return nil, nil
}

// GetItemsByUser - return list of items by user
func (m *MemoryAuction) GetItemsByUser(userID UserID) ([]*Item, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	var out []*Item

	for _, bids := range m.bids {
		for _, b := range bids.bids {
			if b.UserID == userID {
				out = append(out, bids.item)
				break
			}
		}
	}
	return out, nil
}

// AddItem - add new item to auction
func (m *MemoryAuction) AddItem(item *Item) (*Item, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	bids, ok := m.bids[item.ID]
	if !ok {
		bids = &itemBids{item: item}
		m.bids[item.ID] = bids
	}
	return bids.item, nil
}
