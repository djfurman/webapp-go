package main

import (
	"net/http"
	"strconv"

	"github.com/djfurman/webapp-go/internal/cards"
	"github.com/djfurman/webapp-go/internal/models"
	"github.com/go-chi/chi/v5"
)

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "home", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "terminal", &templateData{}, "stripe-js"); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) Receipt(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// Read the posted data
	customerFirstName := r.Form.Get("customer-first-name")
	customerLastName := r.Form.Get("customer-last-name")
	email := r.Form.Get("cardholder-email")
	paymentIntent := r.Form.Get("payment-intent")
	paymentMethod := r.Form.Get("payment-method")
	paymentAmount := r.Form.Get("payment-amount")
	paymentCurrency := r.Form.Get("payment-currency")
	widgetID, _ := strconv.Atoi(r.Form.Get("product-id"))

	card := cards.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.key,
	}

	pi, err := card.RetrievePaymentIntent(paymentIntent)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear

	// Create a new customer
	customerID, err := app.SaveCustomer(customerFirstName, customerLastName, email)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.infoLog.Println(customerID)

	// Create a new transaction
	amount, _ := strconv.Atoi(paymentAmount)
	txn := models.Transaction{
		Amount:              amount,
		Currency:            paymentCurrency,
		LastFour:            lastFour,
		ExpiryMonth:         int(expiryMonth),
		ExpiryYear:          int(expiryYear),
		BankReturnCode:      pi.Charges.Data[0].ID,
		TransactionStatusID: 2,
	}
	txnID, err := app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.infoLog.Println(txnID)

	// Create a new order
	order := models.Order{
		Amount:        amount,
		CustomerID:    customerID,
		Quantity:      1,
		StatusID:      1,
		TransactionID: txnID,
		WidgetID:      widgetID,
	}

	_, err = app.SaveOrder(order)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["email"] = email
	data["customer_first_name"] = customerFirstName
	data["customer_last_name"] = customerLastName
	data["pi"] = paymentIntent
	data["pm"] = paymentMethod
	data["pa"] = paymentAmount
	data["pc"] = paymentCurrency
	data["last_four"] = lastFour
	data["expiry_month"] = expiryMonth
	data["expiry_year"] = expiryYear
	data["bank_return_code"] = pi.Charges.Data[0].ID

	// Should write this data to session, then direct user to a new page

	if err := app.renderTemplate(w, r, "receipt", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

// SaveCustomer saves a new customer to the database and returns an ID
func (app *application) SaveCustomer(firstName, lastName, email string) (int, error) {
	customer := models.Customer{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}

	id, err := app.DB.InsertCustomer(customer)
	if err != nil {
		return 0, err
	}

	return id, err
}

// SaveTransaction saves a new transaction to the database and return an ID
func (app *application) SaveTransaction(txn models.Transaction) (int, error) {
	id, err := app.DB.InsertTransaction(txn)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// SaveOrder saves a new order to the database and returns its ID
func (app *application) SaveOrder(order models.Order) (int, error) {
	orderID, err := app.DB.InsertOrder(order)
	if err != nil {
		return 0, err
	}

	return orderID, nil
}

// ChargeOnce displays the page to buy a single widget
func (app *application) ChargeOnce(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	widgetID, _ := strconv.Atoi(id)

	widget, err := app.DB.GetWidget(widgetID)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["widget"] = widget

	if err := app.renderTemplate(w, r, "buy-once", &templateData{
		Data: data,
	}, "stripe-js"); err != nil {
		app.errorLog.Println(err)
	}
}
