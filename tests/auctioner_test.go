package auctionbidtrackertest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"runtime/debug"
	"strconv"
	"testing"

	"github.com/vkd/auctionbidtracker"
	"github.com/vkd/auctionbidtracker/server"
)

type Auctioner = auctionbidtracker.Auctioner
type Item = auctionbidtracker.Item
type User = auctionbidtracker.User
type ItemID = auctionbidtracker.ItemID
type UserID = auctionbidtracker.UserID
type Bid = auctionbidtracker.Bid

func TestAuctioner(t *testing.T) {
	for _, a := range []Auctioner{
		NewClientAuctioner(),
		auctionbidtracker.NewMemoryAuction(),
	} {
		if a == nil {
			t.Errorf("Auctioner is nil")
			continue
		}

		item1 := addItem(t, a)

		user1 := &User{ID: 1}

		makeBid(t, a, item1, user1, 2)
		makeBid(t, a, item1, user1, 3)
		checkWinning(t, a, item1, 3)

		user2 := &User{ID: 2}
		makeBid(t, a, item1, user2, 1)
		checkWinning(t, a, item1, 3)

		makeBid(t, a, item1, user2, 4)
		checkWinning(t, a, item1, 4)

		item2 := addItem(t, a)
		makeBid(t, a, item2, user2, 5)
		checkWinning(t, a, item1, 4)
		checkWinning(t, a, item2, 5)

		item3 := addItem(t, a)
		makeBid(t, a, item3, user2, 7)

		checkAllBids(t, a, item1, []int{2, 3, 1, 4})
		checkAllBids(t, a, item2, []int{5})
		checkAllBids(t, a, item3, []int{7})

		checkItemsByUser(t, a, user1, []*Item{item1, item2})
		checkItemsByUser(t, a, user2, []*Item{item1, item2, item3})
	}
}

func mustError(t *testing.T, err error) {
	if err != nil {
		debug.PrintStack()
		t.Fatalf(err.Error())
	}
}

var nextItemID ItemID = 1

func addItem(t *testing.T, a Auctioner) *Item {
	item, err := a.AddItem(&Item{ID: nextItemID})
	mustError(t, err)
	nextItemID++
	return item
}

func makeBid(t *testing.T, a Auctioner, item *Item, user *User, value int) {
	mustError(t, a.MakeBid(item.ID, user.ID, &Bid{Value: value}))
}

func checkWinning(t *testing.T, a Auctioner, item *Item, value int) {
	b, err := a.GetWinningBid(item.ID)
	mustError(t, err)
	if b.Value != value {
		debug.PrintStack()
		t.Fatalf("Wrong winning bid: %v", b)
	}
}

func checkAllBids(t *testing.T, a Auctioner, item *Item, values []int) {
	bb, err := a.GetAllBids(item.ID)
	mustError(t, err)
	for i, b := range bb {
		if b.Value != values[i] {
			t.Errorf("Wrong get all bids: %v", bb)
		}
	}
}

func checkItemsByUser(t *testing.T, a Auctioner, user *User, items []*Item) {
	ii, err := a.GetItemsByUser(user.ID)
	mustError(t, err)

	for i, item := range ii {
		if item.ID != items[i].ID {
			t.Errorf("Wrong get items by user: %v", ii)
		}
	}
}

type clientAuctioner struct {
	handler http.Handler
}

func NewClientAuctioner() *clientAuctioner {
	return &clientAuctioner{
		handler: server.NewServer(auctionbidtracker.NewMemoryAuction()),
	}
}

// var testAddr = "http://localhost:8080"

var _ Auctioner = (*clientAuctioner)(nil)

func (c *clientAuctioner) MakeBid(itemID ItemID, userID UserID, bid *Bid) error {
	body, err := json.Marshal(&bid)
	if err != nil {
		return err
	}

	resp := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/items/"+strconv.Itoa(itemID.Int())+"/bids?user_id="+strconv.Itoa(int(userID)), bytes.NewReader(body))
	c.handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("wrong response status code: %d %s", resp.Code, string(body))
	}
	var j struct {
		Status string `json:"status"`
	}
	err = unmarshalBody(resp.Body, &j)
	if err != nil {
		return err
	}
	if j.Status != "ok" {
		return fmt.Errorf("wrong status: %q", j.Status)
	}
	return nil
}

func (c *clientAuctioner) GetWinningBid(itemID ItemID) (*Bid, error) {
	resp := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/items/"+strconv.Itoa(itemID.Int())+"/winning-bid", nil)
	c.handler.ServeHTTP(resp, req)

	var j Bid
	err := unmarshalBody(resp.Body, &j)
	if err != nil {
		return nil, err
	}
	return &j, nil
}
func (c *clientAuctioner) GetAllBids(itemID ItemID) ([]*Bid, error) {
	resp := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/items/"+strconv.Itoa(itemID.Int())+"/bids", nil)
	c.handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("wrong response status code: %d %s", resp.Code, string(body))
	}
	var bids []*Bid
	err := unmarshalBody(resp.Body, &bids)
	if err != nil {
		return nil, err
	}
	return bids, nil
}
func (c *clientAuctioner) GetItemsByUser(userID UserID) ([]*Item, error) {
	resp := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/items?user_id="+strconv.Itoa(int(userID)), nil)
	c.handler.ServeHTTP(resp, req)

	var items []*Item
	err := unmarshalBody(resp.Body, &items)
	if err != nil {
		return nil, err
	}
	return items, nil
}
func (c *clientAuctioner) AddItem(item *Item) (*Item, error) {
	body, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}

	resp := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/items", bytes.NewReader(body))
	c.handler.ServeHTTP(resp, req)

	var j Item
	err = unmarshalBody(resp.Body, &j)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

func unmarshalBody(rc io.Reader, j interface{}) error {
	body, err := ioutil.ReadAll(rc)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, j)
	if err != nil {
		return err
	}
	return nil
}
