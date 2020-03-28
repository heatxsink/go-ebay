package ebay

import (
	"encoding/xml"
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/heatxsink/go-httprequest"
)

const (
	GLOBAL_ID_EBAY_US = "EBAY-US"
	GLOBAL_ID_EBAY_FR = "EBAY-FR"
	GLOBAL_ID_EBAY_DE = "EBAY-DE"
	GLOBAL_ID_EBAY_IT = "EBAY-IT"
	GLOBAL_ID_EBAY_ES = "EBAY-ES"
)

// Item for sale on EBay
type Item struct {
	ItemID        string    `xml:"itemId"`
	Title         string    `xml:"title"`
	Location      string    `xml:"location"`
	CurrentPrice  float64   `xml:"sellingStatus>currentPrice"`
	ShippingPrice float64   `xml:"shippingInfo>shippingServiceCost"`
	BinPrice      float64   `xml:"listingInfo>buyItNowPrice"`
	ShipsTo       []string  `xml:"shippingInfo>shipToLocations"`
	ListingURL    string    `xml:"viewItemURL"`
	ImageURL      string    `xml:"galleryURL"`
	Site          string    `xml:"globalId"`
	EndTime       time.Time `xml:"listingInfo>endTime"`
}

// FindItemsResponse from EBay
type FindItemsResponse struct {
	XMLName   xml.Name `xml:"findItemsByKeywordsResponse"`
	Items     []Item   `xml:"searchResult>item"`
	Timestamp string   `xml:"timestamp"`
}

// FindCompletedItemsResponse from EBay
type FindCompletedItemsResponse struct {
	XMLName   xml.Name `xml:"findCompletedItemsResponse"`
	Items     []Item   `xml:"searchResult>item"`
	Timestamp string   `xml:"timestamp"`
}

// ErrorMessage from EBay
type ErrorMessage struct {
	XMLName xml.Name `xml:"errorMessage"`
	Error   Error    `xml:"error"`
}

// Error response from EBay
type Error struct {
	ErrorID   string `xml:"errorId"`
	Domain    string `xml:"domain"`
	Severity  string `xml:"severity"`
	Category  string `xml:"category"`
	Message   string `xml:"message"`
	SubDomain string `xml:"subdomain"`
}

// EBay API request
type EBay struct {
	ApplicationID string
	HTTPRequest   *httprequest.HttpRequest
}

type soldURL func(string, string, int) (string, error)
type searchURL func(string, string, int, bool) (string, error)

// New EBay API request
func New(applicationID string) *EBay {
	e := EBay{}
	e.ApplicationID = applicationID
	e.HTTPRequest = httprequest.NewWithDefaults()
	return &e
}

func (e *EBay) buildSoldURL(globalID string, keywords string, entriesPerPage int) (string, error) {
	filters := url.Values{}
	filters.Add("itemFilter(0).name", "Condition")
	filters.Add("itemFilter(0).value(0)", "Used")
	filters.Add("itemFilter(0).value(1)", "Unspecified")
	filters.Add("itemFilter(1).name", "SoldItemsOnly")
	filters.Add("itemFilter(1).value(0)", "true")
	return e.buildURL(globalID, keywords, "findCompletedItems", entriesPerPage, filters)
}

func (e *EBay) buildSearchURL(globalID string, keywords string, entriesPerPage int, binOnly bool) (string, error) {
	filters := url.Values{}
	filters.Add("itemFilter(0).name", "ListingType")
	filters.Add("itemFilter(0).value(0)", "AuctionWithBIN")

	if !binOnly {
		filters.Add("itemFilter(0).value(1)", "FixedPrice")
		filters.Add("itemFilter(0).value(2)", "Auction")
	}
	return e.buildURL(globalID, keywords, "findItemsByKeywords", entriesPerPage, filters)
}

func (e *EBay) buildURL(globalID string, keywords string, operationName string, entriesPerPage int, filters url.Values) (string, error) {
	var u *url.URL
	u, err := url.Parse("http://svcs.ebay.com/services/search/FindingService/v1")
	if err != nil {
		return "", err
	}
	params := url.Values{}
	params.Add("OPERATION-NAME", operationName)
	params.Add("SERVICE-VERSION", "1.0.0")
	params.Add("SECURITY-APPNAME", e.ApplicationID)
	params.Add("GLOBAL-ID", globalID)
	params.Add("RESPONSE-DATA-FORMAT", "XML")
	params.Add("REST-PAYLOAD", "")
	params.Add("keywords", keywords)
	params.Add("paginationInput.entriesPerPage", strconv.Itoa(entriesPerPage))
	for key := range filters {
		for _, val := range filters[key] {
			params.Add(key, val)
		}
	}
	u.RawQuery = params.Encode()
	return u.String(), err
}

func (e *EBay) findItems(globalID string, keywords string, entriesPerPage int, url string) (FindItemsResponse, error) {
	var response FindItemsResponse
	headers := make(map[string]string)
	headers["User-Agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_3) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11"
	body, statusCode, err := e.HTTPRequest.Get(url, headers)
	if err != nil {
		return response, err
	}
	if statusCode != 200 {
		var em ErrorMessage
		err = xml.Unmarshal([]byte(body), &em)
		if err != nil {
			return response, err
		}
		return response, errors.New(em.Error.Message)
	}
	err = xml.Unmarshal([]byte(body), &response)
	if err != nil {
		return response, err
	}

	return response, err
}

// FindItemsByKeywords returns items matching the keyword search terms
func (e *EBay) FindItemsByKeywords(globalID string, keywords string, entriesPerPage int, binOnly bool) (FindItemsResponse, error) {
	var response FindItemsResponse
	url, err := e.buildSearchURL(globalID, keywords, entriesPerPage, binOnly)
	if err != nil {
		var response FindItemsResponse
		return response, err
	}
	headers := make(map[string]string)
	headers["User-Agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_3) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11"
	body, statusCode, err := e.HTTPRequest.Get(url, headers)
	if err != nil {
		return response, err
	}
	if statusCode != 200 {
		var em ErrorMessage
		err = xml.Unmarshal([]byte(body), &em)
		if err != nil {
			return response, err
		}
		return response, errors.New(em.Error.Message)
	}
	err = xml.Unmarshal([]byte(body), &response)
	if err != nil {
		return response, err
	}
	return response, err
}

// FindSoldItems returns sold items by keyword
func (e *EBay) FindSoldItems(globalID string, keywords string, entriesPerPage int) (FindCompletedItemsResponse, error) {
	var response FindCompletedItemsResponse
	url, err := e.buildSoldURL(globalID, keywords, entriesPerPage)
	if err != nil {
		return response, err
	}
	headers := make(map[string]string)
	headers["User-Agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_3) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11"
	body, statusCode, err := e.HTTPRequest.Get(url, headers)
	if err != nil {
		return response, err
	}
	if statusCode != 200 {
		var em ErrorMessage
		err = xml.Unmarshal([]byte(body), &em)
		if err != nil {
			return response, err
		}
		return response, errors.New(em.Error.Message)
	}
	err = xml.Unmarshal([]byte(body), &response)
	if err != nil {
		return response, err
	}

	return response, err
}
