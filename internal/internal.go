package internal

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/seanomeara96/go-bigcommerce"
)

type JobType = int

const CaterHireJobType JobType = 1
const HireAlljobType JobType = 2

func jobTypeToWebsiteName(jobType JobType) string {
	if jobType == 1 {
		return string(CATERHIRE)
	} else if jobType == 2 {
		return string(HIREALL)
	}
	return ""
}

type Delivery int

const DELIVERY Delivery = 0
const COLLECTION Delivery = 1

type Orders struct {
	XMLName xml.Name `xml:"Orders"`
	Orders  []Order  `xml:"Order"`
}

type Address struct {
	Street1 string `xml:"Street1"`
	Street2 string `xml:"Street2"`
	City    string `xml:"City"`
	State   string `xml:"State"`
	Zip     string `xml:"Zip"`
}

type Order struct {
	JobType              JobType        `xml:"JobType"`
	WebEnquiryID         string         `xml:"webenquiryid"`
	FirstContactDate     string         `xml:"FirstContactDate"`
	Name                 string         `xml:"Name"`
	BillingCompany       string         `xml:"BillingCompany"`
	BillingStreet1       string         `xml:"BillingStreet1"`
	BillingStreet2       string         `xml:"BillingStreet2"`
	BillingCity          string         `xml:"BillingCity"`
	BillingState         string         `xml:"BillingState"`
	BillingZip           string         `xml:"BillingZip"`
	Email                string         `xml:"Email"`
	TelNo                string         `xml:"TelNo"`
	DeliveryType         Delivery       `xml:"DeliveryType"`
	DeliveryName         string         `xml:"Deliveryname"`
	DeliveryCompany      string         `xml:"DeliveryCompany"`
	DeliveryStreet1      string         `xml:"DeliveryStreet1"`
	DeliveryStreet2      string         `xml:"DeliveryStreet2"`
	DeliveryCity         string         `xml:"DeliveryCity"`
	DeliveryState        string         `xml:"DeliveryState"`
	DeliveryZip          string         `xml:"DeliveryZip"`
	DeliveryInstructions string         `xml:"Deliveryinstructions"`
	DeliveryDate         string         `xml:"DeliveryDate"`
	CollectionDate       string         `xml:"CollectionDate"`
	ShippingTotal        string         `xml:"ShippingTotal"`
	OrderLineItems       OrderLineItems `xml:"OrderLineItems"`
	OtherInfo            string         `xml:"OtherInfo"`
}

func (o Order) Validate() error {
	if utf8.RuneCountInString(o.DeliveryInstructions) > 512 {
		return fmt.Errorf("too many characters in order comment")
	}

	if o.JobType != CaterHireJobType && o.JobType != HireAlljobType {
		return fmt.Errorf("expected jobtype code %d or %d. got %d instead", CaterHireJobType, HireAlljobType, o.JobType)
	}

	return nil
}

type OrderLineItems struct {
	Items []OrderLineItem `xml:"OrderLineItem"`
}

type OrderLineItem struct {
	ID       string  `xml:"Id"`
	Name     string  `xml:"Name"`
	SKU      string  `xml:"SKU"`
	Quantity int     `xml:"Quantity"`
	Price    float64 `xml:"Price"`
	Subtotal float64 `xml:"Subtotal"`
}

func ConvertOrderProductToItem(op bigcommerce.OrderProduct) (OrderLineItem, error) {
	price, err := strconv.ParseFloat(op.BasePrice, 64)
	if err != nil {
		return OrderLineItem{}, fmt.Errorf("error parsing BasePrice: %w", err)
	}
	subtotal, err := strconv.ParseFloat(op.TotalExTax, 64)
	if err != nil {
		return OrderLineItem{}, fmt.Errorf("error parsing TotalExTax: %w", err)
	}

	return OrderLineItem{
		ID:       strconv.Itoa(op.ID),
		Name:     op.Name,
		SKU:      op.SKU,
		Quantity: op.Quantity,
		Price:    price,
		Subtotal: subtotal,
	}, nil
}

// removePatterns removes all substrings enclosed in /**...**/ or /*/.../*/ from the input message.
func removeComments(message string) string {
	// Regular expression to match /**...**/ or /*/.../*/
	re := regexp.MustCompile(`/\*.*?\*/.*/\*.*?\*/`)
	return re.ReplaceAllString(message, "")
}

func extractComments(message string) string {
	// Regular expression to match /**...**/ or /*/.../*/
	re := regexp.MustCompile(`/\*.*?\*/.*/\*.*?\*/`)
	return re.FindString(message)
}

func ConvertOrderToHireJob(startDate, endDate string, order bigcommerce.Order, deliveryType Delivery, shippingAddress bigcommerce.ShippingAddress, orderProducts []bigcommerce.OrderProduct) (Order, error) {
	var items []OrderLineItem
	for _, p := range orderProducts {
		item, err := ConvertOrderProductToItem(p)
		if err != nil {
			return Order{}, err
		}
		items = append(items, item)
	}

	billingName := fmt.Sprintf("%s %s", order.BillingAddress.FirstName, order.BillingAddress.LastName)
	billingAddress := Address{order.BillingAddress.Street1, order.BillingAddress.Street2, order.BillingAddress.City, order.BillingAddress.State, order.BillingAddress.Zip}
	shippingName := fmt.Sprintf("%s %s", shippingAddress.FirstName, shippingAddress.LastName)
	deliveryAddress := Address{shippingAddress.Street1, shippingAddress.Street2, shippingAddress.City, shippingAddress.State, shippingAddress.Zip}

	deliveryInstructions := strings.TrimSpace(removeComments(order.CustomerMessage))
	otherInfo := strings.TrimSpace(extractComments(order.CustomerMessage))

	return Order{
		WebEnquiryID:         fmt.Sprintf("%d", order.ID),
		FirstContactDate:     startDate, // Assuming FirstContactDate is the same as StartDate
		Name:                 billingName,
		BillingCompany:       order.BillingAddress.Company,
		BillingStreet1:       billingAddress.Street1,
		BillingStreet2:       billingAddress.Street2,
		BillingCity:          billingAddress.City,
		BillingState:         billingAddress.State,
		BillingZip:           billingAddress.Zip,
		Email:                order.BillingAddress.Email,
		TelNo:                order.BillingAddress.Phone,
		DeliveryName:         shippingName,
		DeliveryCompany:      shippingAddress.Company,
		DeliveryStreet1:      deliveryAddress.Street1,
		DeliveryStreet2:      deliveryAddress.Street2,
		DeliveryCity:         deliveryAddress.City,
		DeliveryState:        deliveryAddress.State,
		DeliveryZip:          deliveryAddress.Zip,
		DeliveryInstructions: deliveryInstructions,
		OtherInfo:            otherInfo,
		DeliveryDate:         startDate,
		CollectionDate:       endDate,
		ShippingTotal:        order.ShippingCostExTax,
		OrderLineItems:       OrderLineItems{Items: items},
		DeliveryType:         deliveryType,
	}, nil
}

func extractDatesFromCustomerMessage(customerMessage string) (startDate string, endDate string, err error) {
	dateRegexps := map[string]*regexp.Regexp{
		"delivery":   regexp.MustCompile(`Delivery\sDate\s=\s(\w+,\s\w+\s\d{1,2},\s\d{4})`),
		"collection": regexp.MustCompile(`Collection\sDate\s=\s(\w+,\s\w+\s\d{1,2},\s\d{4})`),
		"pickup":     regexp.MustCompile(`Pickup\sDate\s=\s(\w+,\s\w+\s\d{1,2},\s\d{4})`),
		"return":     regexp.MustCompile(`Pickup\sperson\s=\s(\w+,\s\w+\s\d{1,2},\s\d{4})`),
	}

	startDate, endDate = "", ""
	for key, re := range dateRegexps {
		matches := re.FindStringSubmatch(customerMessage)
		if len(matches) > 1 {
			if key == "delivery" || key == "pickup" {
				startDate = matches[1]
			} else if key == "collection" || key == "return" {
				endDate = matches[1]
			}
		}
	}

	if startDate == "" || endDate == "" {
		return "", "", fmt.Errorf("could not extract dates from customer message: %s", customerMessage)
	}

	_startDate, err := time.Parse("Monday, January 2, 2006", startDate)
	if err != nil {
		return "", "", err
	}
	_endDate, err := time.Parse("Monday, January 2, 2006", endDate)
	if err != nil {
		return "", "", err
	}

	startDate = _startDate.Format("02-01-2006")
	endDate = _endDate.Format("02-01-2006")

	return startDate, endDate, nil
}

func xmlToFile(fileName string, data []byte) error {
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating file %s: %v", fileName, err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("error writing to file %s: %v", fileName, err)
	}

	return nil
}

func orderToXML(client *bigcommerce.Client, jobType JobType, order bigcommerce.Order) ([]byte, error) {

	startDate, endDate := "", ""
	integerStringExp := regexp.MustCompile(`\*\/(.+);\/\*`)
	matches := integerStringExp.FindStringSubmatch(order.CustomerMessage)
	if len(matches) < 2 {
		log.Printf("[WARNING] no match found for order %d customer message: %s", order.ID, order.CustomerMessage)
	} else {
		integerString := matches[1]
		start, end, err := extractDatesFromCustomerMessage(integerString)
		if err != nil {
			return nil, fmt.Errorf("error extracting dates for order %d: %v", order.ID, err)
		}
		startDate, endDate = start, end
	}

	var (
		page     = 1
		limit    = 50
		products = []bigcommerce.OrderProduct{}
	)
	for {
		batch, _, err := client.V2.GetOrderProducts(order.ID, bigcommerce.OrderProductsQueryParams{Page: page, Limit: limit})
		if err != nil {
			return nil, fmt.Errorf("error getting order products for order %d: %v", order.ID, err)
		}
		products = append(products, batch...)
		if len(batch) < limit {
			break
		}
		page++
	}

	deliveryType := DELIVERY

	shippingCost, err := strconv.ParseFloat(order.ShippingCostExTax, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse shipping cost float %s: %v", order.ShippingCostExTax, err)
	}

	shippingAddresses, err := client.V2.GetOrderShippingAddress(order.ID, bigcommerce.ShippingAddressQueryParams{})
	if err != nil {
		return nil, fmt.Errorf("error getting shipping addresses for order %d: %v", order.ID, err)
	}

	if len(shippingAddresses) == 0 {
		return nil, fmt.Errorf("no shipping addresses found for order %d", order.ID)
	}

	shippingAddress := shippingAddresses[0]

	if shippingAddress.ShippingMethod != "Flat Rate for Delivery & Collection" && shippingCost == 0.00 {
		deliveryType = COLLECTION
	}

	hireJob, err := ConvertOrderToHireJob(startDate, endDate, order, deliveryType, shippingAddress, products)
	if err != nil {
		return nil, fmt.Errorf("error converting order %d to hire job: %v", order.ID, err)
	}

	hireJob.JobType = jobType
	if err := hireJob.Validate(); err != nil {
		return nil, err
	}

	var orders Orders
	orders.Orders = append(orders.Orders, hireJob)
	b, err := xml.MarshalIndent(orders, "", "    ")
	if err != nil {
		log.Fatalf("error marshalling all orders to XML: %v", err)
	}

	return b, nil
}

type GenerateFilesConfig struct {
	JobType    JobType
	StoreHash  string
	AuthToken  string
	MinOrderID int
}

func GenerateFiles(db *sql.DB, fileDestination string, config GenerateFilesConfig) error {
	client := bigcommerce.NewClient(config.StoreHash, config.AuthToken, nil, nil)
	statuses, err := client.V2.GetOrderStatuses()
	if err != nil {
		return fmt.Errorf("[ERROR] getting order statuses: %v", err)
	}

	statusID := 11
	for _, s := range statuses {
		if s.Name == "Awaiting Fulfillment" {
			statusID = s.ID
			break
		}
	}

	orderSortParams := bigcommerce.OrderSortQuery{
		Field:     bigcommerce.OrderSortFieldID,
		Direction: bigcommerce.OrderSortDirectionDesc,
	}

	orderQueryParams := bigcommerce.OrderQueryParams{
		Limit:    10,
		Sort:     orderSortParams.String(),
		MinID:    config.MinOrderID,
		StatusID: statusID,
	}

	orders, _, err := client.V2.GetOrders(orderQueryParams)
	if err != nil {
		return fmt.Errorf("[ERROR] getting orders: %v", err)
	}

	stmt, err := db.Prepare(`INSERT INTO orders(order_id, xml_file_created, website) VALUES (?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, order := range orders {
		xml, err := orderToXML(client, config.JobType, order)
		if err != nil {
			log.Printf("[ERROR]  %v\n", err)
			continue
		}

		fileName := fileDestination + "order" + strconv.Itoa(order.ID) + ".xml"
		err = xmlToFile(fileName, xml)
		if err != nil {
			return fmt.Errorf("[ERROR] %v", err)
		}

		_, err = stmt.Exec(order.ID, time.Now().UTC(), jobTypeToWebsiteName(config.JobType))
		if err != nil {
			return err
		}

	}
	return nil
}
