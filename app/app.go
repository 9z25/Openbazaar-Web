package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	jwt "github.com/dgrijalva/jwt-go"
)

var mySigningKey = []byte("captainjacksparrowsayshi")

const (
	STATIC_DIR = "/public/"
	password   = "Basic dGVzdC11c2VyOnRlc3QtdXNlcg=="
)

//App is a struct
type App struct {
	Router   *mux.Router
	Database *sql.DB
}

type dataTable struct {
	Headers        map[int]string
	PurchaseOrders []map[int]string
}

type addresses struct {
	BCH string `json:"BCH"`
	BTC string `json:"BTC"`
	LTC string `json:"LTC"`
	ZEC string `json:"ZEC"`
}

type output struct {
	Address string `json:"address"`
	Value   int64  `json:"value"`
	Index   int    `json:"index"`
	OrderID string `json:"orderid"`
}

type amount struct {
	Amount string `json:"amount"`
}

type purchaseOrder struct {
	ShipTo       string `json:"shipTo"`
	EmailAddress string `json:"emailAddress"`
	Address      string `json:"address"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
	PostalCode   string `json:"postalCode"`
	Amount       string `json:"amount"`
}

type paymentKeys struct {
	OrderID        string `json:"orderID"`
	PaymentAddress string `json:"paymentAddress"`
}

var m addresses
var n bool
var price amount

func (app *App) purchaseOrderPost(w http.ResponseWriter, r *http.Request) {
	var po purchaseOrder

	d, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(d, &po)
	//get BCH address
	var url = "http://localhost:4002/wallet/newaddress"
	req, _ := http.NewRequest("GET", url, nil)

	//Should use token
	req.Header.Set("Authorization", password)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)

	json.Unmarshal(data, &m)
	n = false
	oid := uuid.New().String()

	sqlInsert, err := app.Database.Prepare("INSERT INTO PurchaseOrder(email_address, payment_address, order_id,paid, price, ship_to, street_address, city, state, country, postal_code) VALUES(?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		fmt.Println(err.Error())
	}

	//insert into database
	sqlInsert.Exec(po.EmailAddress, m.BCH, oid, 0, po.Amount, po.ShipTo, po.Address, po.City, po.State, po.Country, po.PostalCode)

	//insert into database
	fmt.Fprintf(w, fmt.Sprintf(`{"orderID":"%s","paymentAddress":"%s","emailAddress":"%s"}`, oid, m.BCH, po.EmailAddress))

}

func (app *App) exportPurchaseOrder(w http.ResponseWriter, r *http.Request) {

	sqlQuery, err := app.Database.Query("SELECT * from PurchaseOrder")
	if err != nil {
		panic(err.Error())
	}
	i := 1

	f := excelize.NewFile()
	style, err := f.NewStyle(`{"font":{"bold":true,"italic":false,"family":"Yu Gothic"}}`)
	if err != nil {
		fmt.Println(err)
	}
	f.SetCellValue("Sheet1", "A1", "Number")
	f.SetCellStyle("Sheet1", "A1", "A1", style)
	f.SetCellValue("Sheet1", "B1", "Email Address")
	f.SetCellStyle("Sheet1", "B1", "B1", style)
	f.SetCellValue("Sheet1", "C1", "Payment Address")
	f.SetCellStyle("Sheet1", "C1", "C1", style)
	f.SetCellValue("Sheet1", "D1", "Order ID")
	f.SetCellStyle("Sheet1", "D1", "D1", style)
	f.SetCellValue("Sheet1", "E1", "Paid")
	f.SetCellStyle("Sheet1", "E1", "E1", style)
	f.SetCellValue("Sheet1", "F1", "Price")
	f.SetCellStyle("Sheet1", "F1", "F1", style)
	f.SetCellValue("Sheet1", "G1", "Ship To")
	f.SetCellStyle("Sheet1", "G1", "G1", style)
	f.SetCellValue("Sheet1", "H1", "Shipping Address")
	f.SetCellStyle("Sheet1", "H1", "H1", style)
	f.SetCellValue("Sheet1", "I1", "City")
	f.SetCellStyle("Sheet1", "I1", "i1", style)
	f.SetCellValue("Sheet1", "J1", "State")
	f.SetCellStyle("Sheet1", "j1", "j1", style)
	f.SetCellValue("Sheet1", "K1", "Country")
	f.SetCellStyle("Sheet1", "K1", "K1", style)
	f.SetCellValue("Sheet1", "L1", "Postal Code")
	f.SetCellStyle("Sheet1", "L1", "L1", style)

	for sqlQuery.Next() {
		var number int
		var emailAddress, paymentAddress, orderID, paid, price, shipTo, shippingAddress, city, state, country, postalCode string
		//columns
		if err := sqlQuery.Scan(&number, &emailAddress, &paymentAddress, &orderID, &paid, &price, &shipTo, &shippingAddress, &city, &state, &country, &postalCode); err != nil {
			fmt.Printf("Scan Error: %v\n", err)
		} else {
			i++
			var num = strconv.Itoa(i)
			f.SetCellValue("Sheet1", "A"+num, number)
			f.SetCellValue("Sheet1", "B"+num, emailAddress)
			f.SetCellValue("Sheet1", "C"+num, paymentAddress)
			f.SetCellValue("Sheet1", "D"+num, orderID)
			f.SetCellValue("Sheet1", "E"+num, paid)
			f.SetCellValue("Sheet1", "F"+num, price)
			f.SetCellValue("Sheet1", "G"+num, shipTo)
			f.SetCellValue("Sheet1", "H"+num, shippingAddress)
			f.SetCellValue("Sheet1", "I"+num, city)
			f.SetCellValue("Sheet1", "J"+num, state)
			f.SetCellValue("Sheet1", "K"+num, country)
			f.SetCellValue("Sheet1", "L"+num, postalCode)
		}
	}
	fileName := "purchase_order-" + time.Now().String() + ".xlsx"
	if err := f.SaveAs("./export/" + fileName); err != nil {
		fmt.Println(err)
	}
	fmt.Fprintf(w, `{"progress":"complete","path":./export/`+fileName+`"}`)

}

func (app *App) transactionReceived(w http.ResponseWriter, r *http.Request) {

	sqlSelect := "SELECT payment_address FROM PurchaseOrder"

	d, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	//db query statement
	rows, err := app.Database.Query(sqlSelect)
	if err != nil {
		fmt.Printf("Query : %v\n", err)
	}
	defer rows.Close()

	var pa string
	//var ds, dsr = 1, 1
	var dsr = 1
	for rows.Next() {
		//columns
		if err := rows.Scan(&pa); err != nil {
			fmt.Printf("Scan Error: %v\n", err)
		} else {
			a, _ := regexp.Compile(pa)
			b, _ := regexp.Compile(price.Amount)
			addrMatch := a.MatchString(string(d))
			amountMatch := b.MatchString(string(d))
			if addrMatch == true && amountMatch == true {
				n = true
				//update to true
				sqlUpdate, err := app.Database.Prepare("UPDATE PurchaseOrder SET paid = ? WHERE number = ?")
				if err != nil {
					fmt.Println(err.Error())
				}
				sqlUpdate.Exec(1, dsr)
				return
			}
			dsr++
		}
	}

}

func (app *App) saleComplete(w http.ResponseWriter, r *http.Request) {
	var validToken string
	if n == true {
		validToken, _ = GenerateJWT()
	}
	fmt.Fprintf(w, fmt.Sprintf(`{"bool":"%s","token":"%s"}`, strconv.FormatBool(n), validToken))
}

func (app *App) serveIndex(w http.ResponseWriter, r *http.Request) {
	var number int
	var emailAddress, paymentAddress, orderID, paid, price, shipTo, shippingAddress, city, state, country, postalCode string
	var q = []map[int]string{}
	sqlSelect := "SELECT * FROM PurchaseOrder"

	//db query statement
	rows, err := app.Database.Query(sqlSelect)
	if err != nil {
		fmt.Printf("Query : %v\n", err)
	}
	defer rows.Close()

	for rows.Next() {
		//columns
		if err := rows.Scan(&number, &emailAddress, &paymentAddress, &orderID, &paid, &price, &shipTo, &shippingAddress, &city, &state, &country, &postalCode); err != nil {
			fmt.Printf("Scan Error: %v\n", err)
		} else {
			addRow := map[int]string{1: strconv.Itoa(number), 2: emailAddress, 3: paymentAddress, 4: orderID, 5: paid, 6: price, 7: shipTo, 8: shippingAddress, 9: city, 10: state, 11: country, 12: postalCode}
			q = append(q, addRow)
		}
	}

	var h = map[int]string{1: "Number", 2: "Email Address", 3: "Payment Address", 4: "Order ID", 5: "Paid", 6: "Price",
		7: "Ship To", 8: "Shipping Address", 9: "City", 10: "State", 11: "Country", 12: "Postal Code"}

	data := dataTable{Headers: h}
	data.PurchaseOrders = q

	t, err := template.ParseFiles("./public/export.html")
	if err != nil {
		fmt.Println(err)
	}
	t.Execute(w, data)

}

func (app *App) serveEditor(w http.ResponseWriter, r *http.Request) {

	content, err := ioutil.ReadFile("./public/shopping.html")
	if err != nil {
		log.Fatal(err)
	}

	data := map[string]string{"html": string(content)}
	t, err := template.ParseFiles("./public/editor.html")
	if err != nil {
		fmt.Println(err)
	}
	t.Execute(w, data)
}

func (app *App) serveDownloadPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./public/download.html")
}

func (app *App) serveCatalog(w http.ResponseWriter, r *http.Request) {
	var price int
	var name, currency, shortDesc, desc, image string
	var q = []map[string]string{}
	sqlSelect := "SELECT * FROM Product"

	//db query statement
	rows, err := app.Database.Query(sqlSelect)
	if err != nil {
		fmt.Printf("Query : %v\n", err)
	}
	defer rows.Close()

	for rows.Next() {
		//columns
		if err := rows.Scan(&name, &currency, &price, &shortDesc, &desc, &image); err != nil {
			fmt.Printf("Scan Error: %v\n", err)
		} else {
			priceString := strconv.Itoa(price)
			addRow := map[string]string{"name": name, "currency": currency, "price": priceString, "shortDesc": shortDesc, "desc": desc, "image": image}
			q = append(q, addRow)
		}
	}

	t, err := template.ParseFiles("./public/shopping.html")
	if err != nil {
		fmt.Println(err)
	}
	t.Execute(w, q)

}

func (app *App) serveDownload(w http.ResponseWriter, r *http.Request) {

	fn := make(map[string]string)
	err := json.NewDecoder(r.Body).Decode(&fn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sanitizedName := strings.ReplaceAll(fn["fileName"], "%20", " ")
	fileName := "./product/" + sanitizedName
	fmt.Println(fileName)
	http.ServeFile(w, r, fileName)
}

func (app *App) isAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header["Token"] != nil {

			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("There was an error")
				}
				return mySigningKey, nil
			})

			if err != nil {
				fmt.Fprintf(w, err.Error())
			}

			if token.Valid {
				endpoint(w, r)
			}
		} else {

			fmt.Fprintf(w, "Not Authorized")
		}
	})
}

func GenerateJWT() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["client"] = "Elliot Forbes"
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString(mySigningKey)

	if err != nil {
		fmt.Errorf("Something Went Wrong: %s", err.Error())
		return "", err
	}

	return tokenString, nil
}

//SetupRouter is used for endpoints
func (app *App) SetupRouter() {

	app.Router.
		Methods("POST").
		Path("/api/purchase/").
		HandlerFunc(app.purchaseOrderPost)
	app.Router.
		Methods("GET").
		Path("/api/export/").
		HandlerFunc(app.exportPurchaseOrder)
	app.Router.
		Path("/index/").
		HandlerFunc(app.serveIndex)
	app.Router.
		Path("/editor/").
		HandlerFunc(app.serveEditor)
	app.Router.
		Path("/shopping/").
		HandlerFunc(app.serveCatalog)
	app.Router.
		Path("/download").
		HandlerFunc(app.serveDownloadPage)
	app.Router.
		PathPrefix(STATIC_DIR).
		Handler(http.StripPrefix(STATIC_DIR, http.FileServer(http.Dir("."+STATIC_DIR))))
	app.Router.
		Methods("PUT").
		Path("/api/transactionreceived/").
		HandlerFunc(app.transactionReceived)
	app.Router.
		Methods("GET").
		Path("/api/salecomplete/").
		HandlerFunc(app.saleComplete)
	app.Router.
		Path("/file/").
		Handler(app.isAuthorized(app.serveDownload))
}
