package auctionbidtracker

import (
	"fmt"
	"sync"
)

// FastMemoryAuction - auction with storage in memory
//
// This implementation of auction has two indexes with bids and with items by user where he make a bid
//
// Benchmark_MemoryAuction-4       	    1000	   1750046 ns/op	  484884 B/op	   10135 allocs/op
// Benchmark_FastMemoryAuction-4   	    1000	   1783915 ns/op	  487225 B/op	   10180 allocs/op
//
// Not really fast as I think =(
type FastMemoryAuction struct {
	items *itemsIndex
	users *usersIndex
}

var _ Auctioner = (*FastMemoryAuction)(nil)

// NewFastMemoryAuction - create new memoty auction
func NewFastMemoryAuction() *FastMemoryAuction {
	return &FastMemoryAuction{
		items: newItemsIndex(),
		users: newUsersIndex(),
	}
}

// MakeBid - make new bid for item
func (m *FastMemoryAuction) MakeBid(itemID ItemID, userID UserID, bid *Bid) error {
	bid.UserID = userID

	item, err := m.items.AddBid(itemID, bid)
	if err != nil {
		return err
	}
	return m.users.AddItem(userID, item)
}

// GetWinningBid - get winning bid of item
func (m *FastMemoryAuction) GetWinningBid(itemID ItemID) (*Bid, error) {
	return m.items.GetWinningBid(itemID)
}

// GetAllBids - return list of bids
func (m *FastMemoryAuction) GetAllBids(itemID ItemID) ([]*Bid, error) {
	return m.items.GetBids(itemID)
}

// GetItemsByUser - return list of items by user
func (m *FastMemoryAuction) GetItemsByUser(userID UserID) ([]*Item, error) {
	return m.users.GetItems(userID)
}

// AddItem - add new item to auction
func (m *FastMemoryAuction) AddItem(item *Item) (*Item, error) {
	return m.items.AddItem(item)
}

type itemsIndex struct {
	mx sync.RWMutex

	itemBids map[ItemID]*fastItemBids
}

func newItemsIndex() *itemsIndex {
	return &itemsIndex{
		itemBids: make(map[ItemID]*fastItemBids),
	}
}

func (i *itemsIndex) AddBid(itemID ItemID, bid *Bid) (*Item, error) {
	if bid == nil {
		return nil, fmt.Errorf("bid is nil")
	}
	i.mx.RLock()
	is, ok := i.itemBids[itemID]
	i.mx.RUnlock()

	if !ok {
		return nil, fmt.Errorf("item not found")
	}

	return is.AddBid(bid)
}

func (i *itemsIndex) GetWinningBid(itemID ItemID) (*Bid, error) {
	i.mx.RLock()
	is, ok := i.itemBids[itemID]
	i.mx.RUnlock()

	if !ok {
		return nil, fmt.Errorf("item not found")
	}
	return is.GetWinningBid()
}

func (i *itemsIndex) GetBids(itemID ItemID) ([]*Bid, error) {
	i.mx.RLock()
	is, ok := i.itemBids[itemID]
	i.mx.RUnlock()
	if !ok {
		return nil, fmt.Errorf("item not found")
	}
	return is.GetBids()
}

func (i *itemsIndex) AddItem(item *Item) (*Item, error) {
	i.mx.Lock()
	is, ok := i.itemBids[item.ID]
	if !ok {
		is = newFastItemBids(item)
		i.itemBids[item.ID] = is
	}
	i.mx.Unlock()
	return is.item, nil
}

type fastItemBids struct {
	mx sync.Mutex

	item      *Item
	winnerBid *Bid
	bids      []*Bid
}

func newFastItemBids(item *Item) *fastItemBids {
	return &fastItemBids{
		item: item,
	}
}

func (f *fastItemBids) AddBid(bid *Bid) (*Item, error) {
	f.mx.Lock()
	defer f.mx.Unlock()

	f.bids = append(f.bids, bid)
	if CheckWinner(f.winnerBid, bid) {
		f.winnerBid = bid
	}

	return f.item, nil
}

func (f *fastItemBids) GetWinningBid() (*Bid, error) {
	f.mx.Lock()
	wb := f.winnerBid
	f.mx.Unlock()

	if wb == nil {
		return nil, fmt.Errorf("no winning bid")
	}
	return wb, nil
}

func (f *fastItemBids) GetBids() ([]*Bid, error) {
	f.mx.Lock()
	bs := f.bids
	f.mx.Unlock()

	return bs, nil
}

type usersIndex struct {
	mx sync.Mutex

	userItems map[UserID]*fastUserItems
}

func newUsersIndex() *usersIndex {
	return &usersIndex{
		userItems: make(map[UserID]*fastUserItems),
	}
}

func (u *usersIndex) AddItem(userID UserID, item *Item) error {
	if item == nil {
		return fmt.Errorf("item is nil")
	}

	u.mx.Lock()
	us, ok := u.userItems[userID]
	if !ok {
		us = newFastUserItems()
		u.userItems[userID] = us
	}
	u.mx.Unlock()

	return us.AddItem(item)
}

func (u *usersIndex) GetItems(userID UserID) ([]*Item, error) {
	u.mx.Lock()
	us, ok := u.userItems[userID]
	u.mx.Unlock()

	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return us.GetItems()
}

type fastUserItems struct {
	mx sync.Mutex

	m     map[ItemID]struct{}
	items []*Item
}

func newFastUserItems() *fastUserItems {
	return &fastUserItems{
		m: make(map[ItemID]struct{}),
	}
}

func (f *fastUserItems) AddItem(item *Item) error {
	if item == nil {
		return fmt.Errorf("item is nil")
	}

	f.mx.Lock()
	defer f.mx.Unlock()

	if _, ok := f.m[item.ID]; !ok {
		f.m[item.ID] = struct{}{}
		f.items = append(f.items, item)
	}
	return nil
}

func (f *fastUserItems) GetItems() ([]*Item, error) {
	f.mx.Lock()
	is := f.items
	f.mx.Unlock()
	return is, nil
}
