go-ebay
=======

An ebay search api client in golang.

Example
-------
```go
import (
	"github.com/heatxsink/go-ebay"
)

var (
	ebay_application_id = "your_application_id_here"
)


func main() {
	e := ebay.New(ebay_application_id)
	response, err := e.FindItemsByKeywords(ebay.GLOBAL_ID_EBAY_US, "DJM 900, DJM 850", 10)
	if err != nil {
		fmt.Println("ERROR: ", err)
		os.Exit(1)
	}
	for _, i := range response.Items {
		fmt.Println("Title: ", i.Title)
		fmt.Println("\tListing Url:     ", i.ListingUrl)
		fmt.Println("\tBin Price:       ", i.BinPrice)
		fmt.Println("\tCurrent Price:   ", i.CurrentPrice)
		fmt.Println("\tShipping Price:  ", i.ShippingPrice)
		fmt.Println()
	}
}
```