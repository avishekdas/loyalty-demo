package main

import (
	"errors"
	"fmt"
	"strings"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	"regexp"
)

var logger = shim.NewLogger("LoyaltyChaincode")

//==============================================================================================================================
//	 Participant types - Each participant type is mapped to an integer which we use to compare to the value stored in a
//						 user's eCert
//==============================================================================================================================
//CURRENT WORKAROUND USES ROLES CHANGE WHEN OWN USERS CAN BE CREATED SO THAT IT READ 1, 2, 3, 4, 5
const   AUTHORITY		=  "regulator"
const   HOTEL			=  "hotel"
const   AIRLINES		=  "airlines"
const   CUSTOMER		=  "customer"
const   VENDOR			=  "vendor"

//==============================================================================================================================
//	 Structure Definitions
//==============================================================================================================================
//	Chaincode - A blank struct for use with Shim (A HyperLedger included go file used for get/put state
//				and other HyperLedger functions)
//==============================================================================================================================
type  SimpleChaincode struct {
}

//==============================================================================================================================
//	Customer - Defines the structure for a customer object. JSON on right tells it what JSON fields to map to
//			  that element when reading a JSON object into the struct e.g. JSON make -> Struct Make.
//==============================================================================================================================
type Customer struct {
	CustomerID		string `json:"customerID"`
	Name			string `json:"name"`
	Address			string `json:"address"`
	Cashback		int    `json:"cashback"`
	Email          	string `json:"email"`
	Phone           string `json:"phone"`
	Status	        bool   `json:"status"`
}

//==============================================================================================================================
//	Point of Sales - Defines the structure that holds all the PoS for that have been created.
//				Used as an index when querying all vehicles.
//==============================================================================================================================

type PoS struct {
	PoSID				string `json:"posId"`
	PoSName				string `json:"posName"`
	Status				bool   `json:"status"`
	LoyaltyPercentage	int	   `json:"percentage"`
}

//==============================================================================================================================
//	Items - Items brought.
//==============================================================================================================================

type Item struct {
	ItemID		string `json:"itemId"`
	PoSID		string `json:"posId"`
	ItemName	string `json:"itemName"`
	Price		int	   `json:"price"`
}

//==============================================================================================================================
//	CustomerID Holder - Defines the structure that holds all the customerIDs for Customer that have been created.
//				Used as an index when querying all vehicles.
//==============================================================================================================================

type CustomerID_Holder struct {
	CustomerIDs 	[]string `json:"customerIDs"`
}

//==============================================================================================================================
//	CustomerID Holder - Defines the structure that holds all the customerIDs for Customer that have been created.
//				Used as an index when querying all vehicles.
//==============================================================================================================================

type PoSID_Holder struct {
	PoSIDs 	[]string `json:"posIDs"`
}

//==============================================================================================================================
//	CustomerID Holder - Defines the structure that holds all the customerIDs for Customer that have been created.
//				Used as an index when querying all vehicles.
//==============================================================================================================================

type ItemID_Holder struct {
	ItemIDs 	[]string `json:"itemIDIDs"`
}

//==============================================================================================================================
//	Init Function - Called when the user deploys the chaincode
//==============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	var pointOfSales PoS

	bytes, err := json.Marshal(pointOfSales)

    if err != nil { return nil, errors.New("Error creating pos record") }

	err = stub.PutState("pointOfSales", bytes)

	for i:=0; i < len(args); i=i+2 {
		//t.add_pos(stub, args[i], args[i+1])
	}

	return nil, nil
}

//==============================================================================================================================
//	 General Functions
//==============================================================================================================================
//	 get_ecert - Takes the name passed and calls out to the REST API for HyperLedger to retrieve the ecert
//				 for that user. Returns the ecert as retrived including html encoding.
//==============================================================================================================================
func (t *SimpleChaincode) get_ecert(stub shim.ChaincodeStubInterface, name string) ([]byte, error) {

	ecert, err := stub.GetState(name)

	if err != nil { return nil, errors.New("Couldn't retrieve ecert for user " + name) }

	return ecert, nil
}

//==============================================================================================================================
//	 get_caller - Retrieves the username of the user who invoked the chaincode.
//				  Returns the username as a string.
//==============================================================================================================================

func (t *SimpleChaincode) get_username(stub shim.ChaincodeStubInterface) (string, error) {

    username, err := stub.ReadCertAttribute("username");
	if err != nil { return "", errors.New("Couldn't get attribute 'username'. Error: " + err.Error()) }
	return string(username), nil
}

//==============================================================================================================================
//	 check_affiliation - Takes an ecert as a string, decodes it to remove html encoding then parses it and checks the
// 				  		certificates common name. The affiliation is stored as part of the common name.
//==============================================================================================================================

func (t *SimpleChaincode) check_affiliation(stub shim.ChaincodeStubInterface) (string, error) {
    affiliation, err := stub.ReadCertAttribute("role");
	if err != nil { return "", errors.New("Couldn't get attribute 'role'. Error: " + err.Error()) }
	return string(affiliation), nil

}

//==============================================================================================================================
//	 get_caller_data - Calls the get_ecert and check_role functions and returns the ecert and role for the
//					 name passed.
//==============================================================================================================================

func (t *SimpleChaincode) get_caller_data(stub shim.ChaincodeStubInterface) (string, string, error){

	user, err := t.get_username(stub)

    // if err != nil { return "", "", err }

	// ecert, err := t.get_ecert(stub, user);

    // if err != nil { return "", "", err }

	affiliation, err := t.check_affiliation(stub);

    if err != nil { return "", "", err }

	return user, affiliation, nil
}

//==============================================================================================================================
//	 retrieve_customer - Gets the state of the data at customerID in the ledger then converts it from the stored
//					JSON into the Customer struct for use in the contract. Returns the Vehcile struct.
//					Returns empty v if it errors.
//==============================================================================================================================
func (t *SimpleChaincode) retrieve_customer(stub shim.ChaincodeStubInterface, customerID string) (Customer, error) {

	var v Customer

	bytes, err := stub.GetState(customerID);

	if err != nil {	fmt.Printf("RETRIEVE_CUSTOMER: Failed to invoke Customer_code: %s", err); return v, errors.New("RETRIEVE_CUSTOMER: Error retrieving Customer with customerID = " + customerID) }

	err = json.Unmarshal(bytes, &v);

    if err != nil {	fmt.Printf("RETRIEVE_CUSTOMER: Corrupt Customer record "+string(bytes)+": %s", err); return v, errors.New("RETRIEVE_CUSTOMER: Corrupt Customer record"+string(bytes))	}

	return v, nil
}

//==============================================================================================================================
//	 retrieve_item - Gets the state of the data at itemID in the ledger then converts it from the stored
//					JSON into the Item struct for use in the contract. Returns the Vehcile struct.
//					Returns empty v if it errors.
//==============================================================================================================================
func (t *SimpleChaincode) retrieve_item(stub shim.ChaincodeStubInterface, itemID string) (Item, error) {

	var v Item

	bytes, err := stub.GetState(itemID);

	if err != nil {	fmt.Printf("RETRIEVE_ITEM: Failed to invoke ItemID: %s", err); return v, errors.New("RETRIEVE_ITEM: Error retrieving Item with ItemID = " + itemID) }

	err = json.Unmarshal(bytes, &v);

    if err != nil {	fmt.Printf("RETRIEVE_ITEM: Corrupt Item record "+string(bytes)+": %s", err); return v, errors.New("RETRIEVE_ITEM: Corrupt Item record"+string(bytes))	}

	return v, nil
}

//==============================================================================================================================
//	 retrieve_pos - Gets the state of the data at posID in the ledger then converts it from the stored
//					JSON into the Item struct for use in the contract. Returns the Vehcile struct.
//					Returns empty v if it errors.
//==============================================================================================================================
func (t *SimpleChaincode) retrieve_pos(stub shim.ChaincodeStubInterface, posID string) (PoS, error) {

	var v PoS

	bytes, err := stub.GetState(posID);

	if err != nil {	fmt.Printf("RETRIEVE_PoS: Failed to invoke posID: %s", err); return v, errors.New("RETRIEVE_ITEM: Error retrieving PoS with posID = " + posID) }

	err = json.Unmarshal(bytes, &v);

    if err != nil {	fmt.Printf("RETRIEVE_PoS: Corrupt PoS record "+string(bytes)+": %s", err); return v, errors.New("RETRIEVE_ITEM: Corrupt PoS record"+string(bytes))	}

	return v, nil
}

//==============================================================================================================================
// save_changes - Writes to the ledger the Customer struct passed in a JSON format. Uses the shim file's
//				  method 'PutState'.
//==============================================================================================================================
func (t *SimpleChaincode) save_changes(stub shim.ChaincodeStubInterface, v Customer) (bool, error) {

	bytes, err := json.Marshal(v)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error converting customer record: %s", err); return false, errors.New("Error converting customer record") }

	err = stub.PutState(v.Name, bytes)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error storing customer record: %s", err); return false, errors.New("Error storing customer record") }

	return true, nil
}

//==============================================================================================================================
// save_changes_pos - Writes to the ledger the PoS struct passed in a JSON format. Uses the shim file's
//				  method 'PutState'.
//==============================================================================================================================
func (t *SimpleChaincode) save_changes_pos(stub shim.ChaincodeStubInterface, v PoS) (bool, error) {

	bytes, err := json.Marshal(v)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error converting pos record: %s", err); return false, errors.New("Error converting pos record") }

	err = stub.PutState(v.PoSName, bytes)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error storing pos record: %s", err); return false, errors.New("Error storing pos record") }

	return true, nil
}

//==============================================================================================================================
// save_changes_item - Writes to the ledger the Item struct passed in a JSON format. Uses the shim file's
//				  method 'PutState'.
//==============================================================================================================================
func (t *SimpleChaincode) save_changes_item(stub shim.ChaincodeStubInterface, v Item) (bool, error) {

	bytes, err := json.Marshal(v)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error converting pos record: %s", err); return false, errors.New("Error converting pos record") }

	err = stub.PutState(v.ItemName, bytes)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error storing pos record: %s", err); return false, errors.New("Error storing pos record") }

	return true, nil
}

//==============================================================================================================================
//	 Router Functions
//==============================================================================================================================
//	Invoke - Called on chaincode invoke. Takes a function name passed and calls that function. Converts some
//		  initial arguments passed to other things for use in the called function e.g. name -> ecert
//==============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	caller, caller_affiliation, err := t.get_caller_data(stub)

	if err != nil { return nil, errors.New("Error retrieving caller information")}


	if function == "create_customer" {
        return t.create_customer(stub, caller, caller_affiliation, args[0])
	} else if function == "ping" {
        return t.ping(stub)
    } else { 																	// If the function is not a create then there must be a car so we need to retrieve the customer.
		argPos := 1

		if function == "delete_customer" {										// If its a scrap vehicle then only two arguments are passed (no update value) all others have three arguments and the customerID is expected in the last argument
			argPos = 0
		}

		v, err := t.retrieve_customer(stub, args[argPos])

        if err != nil { fmt.Printf("INVOKE: Error retrieving Customer: %s", err); return nil, errors.New("Error retrieving customer") }


        if strings.Contains(function, "update") == false && function != "delete_customer"    { 									// If the function is not an update or a scrappage it must be a transfer so we need to get the ecert of the recipient.


				if function == "buy_item_by_money" {
					argPos := 2
					i, err := t.retrieve_item(stub, args[argPos])
					if err != nil { fmt.Printf("INVOKE: Error retrieving Item: %s", err); return nil, errors.New("Error retrieving Item") }
					return t.buy_item_by_money(stub, v, i, caller, caller_affiliation)
				} else if  function == "buy_item_by_wallet" {
					argPos := 2
					i, err := t.retrieve_item(stub, args[argPos])
					if err != nil { fmt.Printf("INVOKE: Error retrieving Item: %s", err); return nil, errors.New("Error retrieving Item") }
					return t.buy_item_by_wallet(stub, v,  i, caller, caller_affiliation)
				}

		} else if function == "update_name"  	    { return t.update_name(stub, v, caller, caller_affiliation, args[0]) }
		return nil, errors.New("Function of the name "+ function +" doesn't exist.")

	}
}
//=================================================================================================================================
//	Query - Called on chaincode query. Takes a function name passed and calls that function. Passes the
//  		initial arguments passed are passed on to the called function.
//=================================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	if function == "get_customer_details" {
		if len(args) != 1 { fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("QUERY: Incorrect number of arguments passed") }
		v, err := t.retrieve_customer(stub, args[0])
		if err != nil { fmt.Printf("QUERY: Error retrieving v5c: %s", err); return nil, errors.New("QUERY: Error retrieving v5c "+err.Error()) }
		return t.get_customer_details(stub, v)
	} else if function == "check_unique_customer" {
		return t.check_unique_customer(stub, args[0])
	} else if function == "get_customers" {
		return t.get_customers(stub)
	} else if function == "get_ecert" {
		return t.get_ecert(stub, args[0])
	} else if function == "ping" {
		return t.ping(stub)
	}

	return nil, errors.New("Received unknown function invocation " + function)

}

//=================================================================================================================================
//	 Ping Function
//=================================================================================================================================
//	 Pings the peer to keep the connection alive
//=================================================================================================================================
func (t *SimpleChaincode) ping(stub shim.ChaincodeStubInterface) ([]byte, error) {
	return []byte("Hello, world!"), nil
}

//=================================================================================================================================
//	 Create Function
//=================================================================================================================================
//	 Create Customer - Creates the initial JSON for the vehcile and then saves it to the ledger.
//=================================================================================================================================
func (t *SimpleChaincode) create_customer(stub shim.ChaincodeStubInterface, caller string, caller_affiliation string, customerID string) ([]byte, error) {
	var v Customer

	customerId		:= "\"CustomerID\":\""+customerID+"\", "							// Variables to define the JSON
	name            := "\"Name\":\""+customerID+"\", "
	address			:= "\"Address\":\"UNDEFINED\", "
	cashback		:= "\"Cashback\":0, "
	email			:= "\"Email\":\"UNDEFINED\", "
	phone			:= "\"Phone\":\"UNDEFINED\", "
	status			:= "\"Status\":true "
		
	customer_json := "{"+customerId+name+address+cashback+email+phone+status+"}" 	// Concatenates the variables to create the total JSON object
	matched, err := regexp.Match("^[A-z][A-z][0-9]{7}", []byte(customerID))  				// matched = true if the v5cID passed fits format of two letters followed by seven digits
	if err != nil { fmt.Printf("CREATE_CUSTOMER: Invalid customerID: %s", err); return nil, errors.New("Invalid customerID") }

	if customerID  == "" ||	matched == false {
		fmt.Printf("CREATE_CUSTOMER: Invalid customerID provided");
		return nil, errors.New("Invalid customerID provided")
	}

	err = json.Unmarshal([]byte(customer_json), &v)	// Convert the JSON defined above into a customer object for go
	if err != nil { return nil, errors.New("Invalid JSON object") }
	record, err := stub.GetState(v.CustomerID) 								// If not an error then a record exists so cant create a new car with this V5cID as it must be unique
	if record != nil { return nil, errors.New("Customer already exists") }
	
	_, err  = t.save_changes(stub, v)
	if err != nil { fmt.Printf("CREATE_CUSTOMER: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	bytes, err := stub.GetState("customerID")
	if err != nil { return nil, errors.New("Unable to get customerID") }
	var customerIDs CustomerID_Holder
	err = json.Unmarshal(bytes, &customerIDs)
	if err != nil {	return nil, errors.New("Corrupt Customer record") }
	customerIDs.CustomerIDs = append(customerIDs.CustomerIDs, customerID)
	bytes, err = json.Marshal(customerIDs)
	if err != nil { fmt.Print("Error creating Customer record") }
	err = stub.PutState("customerIDs", bytes)
	if err != nil { return nil, errors.New("Unable to put the state") }
	return nil, nil
}

//=================================================================================================================================
//	 update_name
//=================================================================================================================================
func (t *SimpleChaincode) update_name(stub shim.ChaincodeStubInterface, v Customer, caller string, caller_affiliation string, new_value string) ([]byte, error) {

	if 	v.Status == true {
		v.Name = new_value
	} else {
		return nil, errors.New(fmt.Sprint("Not found"))
	}

	_, err := t.save_changes(stub, v)
	if err != nil { fmt.Printf("UPDATE_NAME: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	return nil, nil
}

//=================================================================================================================================
//	 update_address
//=================================================================================================================================
func (t *SimpleChaincode) update_address(stub shim.ChaincodeStubInterface, v Customer, caller string, caller_affiliation string, new_value string) ([]byte, error) {

	if 	v.Status == true {
		v.Address = new_value
	} else {
		return nil, errors.New(fmt.Sprint("Not found"))
	}

	_, err := t.save_changes(stub, v)
	if err != nil { fmt.Printf("UPDATE_ADDRESS: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	return nil, nil
}

//=================================================================================================================================
//	 update_cashback
//=================================================================================================================================
func (t *SimpleChaincode) update_cashback(stub shim.ChaincodeStubInterface, v Customer, caller string, caller_affiliation string, new_value int) ([]byte, error) {

	if 	v.Status == true {
		v.Cashback = new_value
	} else {
		return nil, errors.New(fmt.Sprint("Not found"))
	}

	_, err := t.save_changes(stub, v)
	if err != nil { fmt.Printf("UPDATE_CASHBACK: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	return nil, nil
}

//=================================================================================================================================
//	 update_email
//=================================================================================================================================
func (t *SimpleChaincode) update_email(stub shim.ChaincodeStubInterface, v Customer, caller string, caller_affiliation string, new_value string) ([]byte, error) {

	if 	v.Status == true {
		v.Email = new_value
	} else {
		return nil, errors.New(fmt.Sprint("Not found"))
	}

	_, err := t.save_changes(stub, v)
	if err != nil { fmt.Printf("UPDATE_EMAIL: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	return nil, nil
}

//=================================================================================================================================
//	 Create PoS - Creates the initial JSON for the vehcile and then saves it to the ledger.
//=================================================================================================================================
func (t *SimpleChaincode) create_pos(stub shim.ChaincodeStubInterface, caller string, caller_affiliation string, posID string) ([]byte, error) {
	var v PoS

	posId		:= "\"PoSID\":\""+posID+"\", "							// Variables to define the JSON
	posName     := "\"PoSName\":0, "
	status			:= "\"Status\":\"true\", "
	percentage		:= "\"LoyaltyPercentage\":\"5%\", "
	
	
	pos_json := "{"+posId+posName+status+percentage+"}" 	// Concatenates the variables to create the total JSON object
	matched, err := regexp.Match("^[A-z][A-z][0-9]{7}", []byte(posId))  				// matched = true if the v5cID passed fits format of two letters followed by seven digits
	if err != nil { fmt.Printf("CREATE_POS: Invalid posID: %s", err); return nil, errors.New("Invalid posID") }

	if posId  == "" ||	matched == false {
		fmt.Printf("CREATE_POS: Invalid posID provided");
		return nil, errors.New("Invalid posID provided")
	}
	
	err = json.Unmarshal([]byte(pos_json), &v)	// Convert the JSON defined above into a PoS object for go
	if err != nil { return nil, errors.New("Invalid JSON object") }
	record, err := stub.GetState(v.PoSID) 								// If not an error then a record exists so cant create a new car with this CustomerID as it must be unique
	if record != nil { return nil, errors.New("POS already exists") }
	
	_, err  = t.save_changes_pos(stub, v)
	if err != nil { fmt.Printf("CREATE_POS: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	bytes, err := stub.GetState("posID")
	if err != nil { return nil, errors.New("Unable to get PoSID") }
	var posIDs PoSID_Holder
	err = json.Unmarshal(bytes, &posIDs)
	if err != nil {	return nil, errors.New("Corrupt PoS record") }
	posIDs.PoSIDs = append(posIDs.PoSIDs, posID)
	bytes, err = json.Marshal(posIDs)
	if err != nil { fmt.Print("Error creating PoS record") }
	err = stub.PutState("posIDs", bytes)
	if err != nil { return nil, errors.New("Unable to put the state") }
	return nil, nil
}

//=================================================================================================================================
//	 update_posname
//=================================================================================================================================
func (t *SimpleChaincode) update_posname(stub shim.ChaincodeStubInterface, v PoS, caller string, caller_affiliation string, new_value string) ([]byte, error) {

	if 	v.Status == true {
		v.PoSName = new_value
	} else {
		return nil, errors.New(fmt.Sprint("Not found"))
	}

	_, err := t.save_changes_pos(stub, v)
	if err != nil { fmt.Printf("UPDATE_POSNAME: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	return nil, nil
}

//=================================================================================================================================
//	 update_percentage
//=================================================================================================================================
func (t *SimpleChaincode) update_percentage(stub shim.ChaincodeStubInterface, v PoS, caller string, caller_affiliation string, new_value int) ([]byte, error) {

	if 	v.Status == true {
		v.LoyaltyPercentage = new_value
	} else {
		return nil, errors.New(fmt.Sprint("Not found"))
	}

	_, err := t.save_changes_pos(stub, v)
	if err != nil { fmt.Printf("UPDATE_PERCENTAGE: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	return nil, nil
}

//=================================================================================================================================
//	 Create PoS - Creates the initial JSON for the vehcile and then saves it to the ledger.
//=================================================================================================================================
func (t *SimpleChaincode) create_item(stub shim.ChaincodeStubInterface, caller string, caller_affiliation string, itemID string) ([]byte, error) {
	var v Item

	itemId		:= "\"ItemID\":\""+itemID+"\", "							// Variables to define the JSON
	posID		:= "\"PoSID\":0, "
	itemName	:= "\"Active\":\"UNDEFINED\", "
	price		:= "\"Price\":\"500\", "
	
	
	item_json := "{"+itemId+posID+itemName+price+"}" 	// Concatenates the variables to create the total JSON object
	matched, err := regexp.Match("^[A-z][A-z][0-9]{7}", []byte(itemId))  				// matched = true if the v5cID passed fits format of two letters followed by seven digits
	if err != nil { fmt.Printf("CREATE_ITEM: Invalid itemID: %s", err); return nil, errors.New("Invalid itemID") }

	if itemId  == "" ||	matched == false {
		fmt.Printf("CREATE_ITEM: Invalid itemID provided");
		return nil, errors.New("Invalid itemID provided")
	}
	
	err = json.Unmarshal([]byte(item_json), &v)	// Convert the JSON defined above into a PoS object for go
	if err != nil { return nil, errors.New("Invalid JSON object") }
	record, err := stub.GetState(v.ItemID) 								// If not an error then a record exists so cant create a new car with this CustomerID as it must be unique
	if record != nil { return nil, errors.New("Item already exists") }
	
	_, err  = t.save_changes_item(stub, v)
	if err != nil { fmt.Printf("CREATE_POS: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	bytes, err := stub.GetState("itemID")
	if err != nil { return nil, errors.New("Unable to get ItemID") }
	var itemIDs ItemID_Holder
	err = json.Unmarshal(bytes, &itemIDs)
	if err != nil {	return nil, errors.New("Corrupt Item record") }
	itemIDs.ItemIDs = append(itemIDs.ItemIDs, itemId)
	bytes, err = json.Marshal(itemIDs)
	if err != nil { fmt.Print("Error creating Item record") }
	err = stub.PutState("itemIDs", bytes)
	if err != nil { return nil, errors.New("Unable to put the state") }
	return nil, nil
}

//=================================================================================================================================
//	 update_item_name
//=================================================================================================================================
func (t *SimpleChaincode) update_item_name(stub shim.ChaincodeStubInterface, v Item, caller string, caller_affiliation string, new_value string) ([]byte, error) {

	v.ItemName = new_value
	
	_, err := t.save_changes_item(stub, v)
	if err != nil { fmt.Printf("UPDATE_ITEM_NAME: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	return nil, nil
}

//=================================================================================================================================
//	 update_posid
//=================================================================================================================================
func (t *SimpleChaincode) update_posid(stub shim.ChaincodeStubInterface, v Item, caller string, caller_affiliation string, new_value string) ([]byte, error) {

	v.PoSID = new_value
	
	_, err := t.save_changes_item(stub, v)
	if err != nil { fmt.Printf("UPDATE_POSID: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	return nil, nil
}

//=================================================================================================================================
//	 update_price
//=================================================================================================================================
func (t *SimpleChaincode) update_price(stub shim.ChaincodeStubInterface, v Item, caller string, caller_affiliation string, new_value int) ([]byte, error) {

	v.Price = new_value
	
	_, err := t.save_changes_item(stub, v)
	if err != nil { fmt.Printf("UPDATE_PRICE: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	return nil, nil
}

//=================================================================================================================================
//	 Read Functions
//=================================================================================================================================
//	 get_customer_details
//=================================================================================================================================
func (t *SimpleChaincode) get_customer_details(stub shim.ChaincodeStubInterface, v Customer) ([]byte, error) {

	bytes, err := json.Marshal(v)
	if err != nil { return nil, errors.New("GET_CUSTOMER_DETAILS: Invalid Customer object") }
	return bytes, nil
}

//=================================================================================================================================
//	 get_customers
//=================================================================================================================================

func (t *SimpleChaincode) get_customers(stub shim.ChaincodeStubInterface) ([]byte, error) {
	bytes, err := stub.GetState("customerIDs")
	if err != nil { return nil, errors.New("Unable to get customerIDs") }
	var customerIDs CustomerID_Holder
	err = json.Unmarshal(bytes, &customerIDs)
	if err != nil {	return nil, errors.New("Corrupt CustomerID_Holder") }
	result := "["
	var temp []byte
	var v Customer

	for _, customer := range customerIDs.CustomerIDs {

		v, err = t.retrieve_customer(stub, customer)

		if err != nil {return nil, errors.New("Failed to retrieve Customer")}

		temp, err = t.get_customer_details(stub, v)

		if err == nil {
			result += string(temp) + ","
		}
	}

	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result)-1] + "]"
	}

	return []byte(result), nil
}

//=================================================================================================================================
//	 check_unique_customer
//=================================================================================================================================
func (t *SimpleChaincode) check_unique_customer(stub shim.ChaincodeStubInterface, customerID string) ([]byte, error) {
	_, err := t.retrieve_customer(stub, customerID)
	if err == nil {
		return []byte("false"), errors.New("Customer is not unique")
	} else {
		return []byte("true"), nil
	}
}

//=================================================================================================================================
//	 Transactions
//=================================================================================================================================

//=================================================================================================================================
//	 buy_item_by_money
//=================================================================================================================================
func (t *SimpleChaincode) buy_item_by_money(stub shim.ChaincodeStubInterface, v Customer, i Item, caller string, caller_affiliation string) ([]byte, error) {

	if v.Status == true {
		p, err := t.retrieve_pos(stub, i.PoSID)
		if err != nil { fmt.Printf("INVOKE: Error retrieving PoS: %s", err); return nil, errors.New("Error retrieving PoS") }
		v.Cashback = v.Cashback + (p.LoyaltyPercentage * i.Price)/100
	} else {												// Otherwise if there is an error
		fmt.Printf("buy_item_by_money: Customer Not Active");
        return nil, errors.New(fmt.Sprintf(" Customer Not Active."))
	}
	
	_, err := t.save_changes(stub, v)						// Write new state
	if err != nil {	fmt.Printf("buy_item_by_money: Error saving changes: %s", err); return nil, errors.New("Error saving changes")	}
	return nil, nil									// We are Done
}

func (t *SimpleChaincode) buy_item_by_wallet(stub shim.ChaincodeStubInterface, v Customer, i Item, caller string, caller_affiliation string) ([]byte, error) {

	if v.Status == true {
		if v.Cashback > i.Price {
			v.Cashback = v.Cashback - i.Price
		} else {
			fmt.Printf("buy_item_by_wallet: Not enough balance");
        	return nil, errors.New(fmt.Sprintf(" Not enough balance."))
		}
	} else {									// Otherwise if there is an error
		fmt.Printf("buy_item_by_wallet: Customer Not Active");
        return nil, errors.New(fmt.Sprintf(" Customer Not Active."))
	}
	_, err := t.save_changes(stub, v)						// Write new state
	if err != nil {	fmt.Printf("buy_item_by_wallet: Error saving changes: %s", err); return nil, errors.New("Error saving changes")	}
	return nil, nil									// We are Done
}

//=================================================================================================================================
//	 Main - main - Starts up the chaincode
//=================================================================================================================================
func main() {

	err := shim.Start(new(SimpleChaincode))
	if err != nil { fmt.Printf("Error starting Chaincode: %s", err) }
}