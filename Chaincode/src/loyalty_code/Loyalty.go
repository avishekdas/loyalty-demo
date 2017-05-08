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
//	CustomerID Holder - Defines the structure that holds all the customerIDs for Customer that have been created.
//				Used as an index when querying all vehicles.
//==============================================================================================================================

type CustomerID_Holder struct {
	Customers 	[]string `json:"customers"`
}

//==============================================================================================================================
//	Init Function - Called when the user deploys the chaincode
//==============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	var customerIDs CustomerID_Holder

	bytes, err := json.Marshal(customerIDs)

    if err != nil { return nil, errors.New("Error creating pos record") }

	err = stub.PutState("customerIDs", bytes)

	for i:=0; i < len(args); i=i+2 {
		//t.add_pos(stub, args[i], args[i+1])
	}

	return nil, nil
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

    if err != nil {	fmt.Printf("RETRIEVE_CUSTOMER: Corrupt Customer record "+string(bytes)+": %s", err); return v, errors.New("RETRIEVE_CUSTOMER: Corrupt Customer record"+string(bytes))}

	return v, nil
}

//==============================================================================================================================
// save_changes - Writes to the ledger the Customer struct passed in a JSON format. Uses the shim file's
//				  method 'PutState'.
//==============================================================================================================================
func (t *SimpleChaincode) save_changes(stub shim.ChaincodeStubInterface, v Customer) (bool, error) {

	bytes, err := json.Marshal(v)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error converting customer record: %s", err); return false, errors.New("Error converting customer record") }

	err = stub.PutState(v.CustomerID, bytes)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error storing customer record: %s", err); return false, errors.New("Error storing customer record") }

	return true, nil
}

//==============================================================================================================================
//	 Router Functions
//==============================================================================================================================
//	Invoke - Called on chaincode invoke. Takes a function name passed and calls that function. Converts some
//		  initial arguments passed to other things for use in the called function e.g. name -> ecert
//==============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	if function == "create_customer" {
        return t.create_customer(stub, args[0])
	} else if function == "ping" {
        return t.ping(stub)
    } else { 																	// If the function is not a create then there must be a car so we need to retrieve the customer.
		v, err := t.retrieve_customer(stub, args[0])

        if err != nil { fmt.Printf("INVOKE: Error retrieving Customer: %s", err); return nil, errors.New("Error retrieving customer") }
        if strings.Contains(function, "update") == false && function != "delete_customer"    {
				// if function == "buy_item_by_money" {
				// 	argPos := 2
				// 	i, err := t.retrieve_item(stub, args[argPos])
				// 	if err != nil { fmt.Printf("INVOKE: Error retrieving Item: %s", err); return nil, errors.New("Error retrieving Item") }
				// 	return t.buy_item_by_money(stub, v, i)
				// } else if  function == "buy_item_by_wallet" {
				// 	argPos := 2
				// 	i, err := t.retrieve_item(stub, args[argPos])
				// 	if err != nil { fmt.Printf("INVOKE: Error retrieving Item: %s", err); return nil, errors.New("Error retrieving Item") }
				// 	return t.buy_item_by_wallet(stub, v,  i)
				// }

		} else if function == "update_name" { return t.update_name(stub, v, args[0]) }
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
func (t *SimpleChaincode) create_customer(stub shim.ChaincodeStubInterface, customerID string) ([]byte, error) {
	var v Customer

	// customerId		:= "\"CustomerID\":\""+customerID+"\", "							// Variables to define the JSON
	// name            := "\"Name\":\""+customerID+"\", "
	// address			:= "\"Address\":\"UNDEFINED\", "
	// cashback			:= "\"Cashback\":0, "
	// email			:= "\"Email\":\"UNDEFINED\", "
	// phone			:= "\"Phone\":\"UNDEFINED\", "
	// status			:= "\"Status\":true"
		
	customer_json := "{\"CustomerID\":\""+customerID+"\",\"Name\":\"test\",\"Address\":\"UNDEFINED\",\"Cashback\":0,\"Email\":\"UNDEFINED\",\"Phone\":\"UNDEFINED\",\"Status\":true}" 	// Concatenates the variables to create the total JSON object
	
	matched, err := regexp.Match("^[A-z][A-z][0-9]{7}", []byte(customerID))  				// matched = true if the v5cID passed fits format of two letters followed by seven digits
	if err != nil { fmt.Printf("CREATE_VEHICLE: Invalid customerID: %s", err); return nil, errors.New("Invalid customerID") }
	if	customerID  == "" || matched == false {
		fmt.Printf("CREATE_CUSTOMER: Invalid customerID provided");
		return nil, errors.New("Invalid customerID provided")
	}

	err = json.Unmarshal([]byte(customer_json), &v)	// Convert the JSON defined above into a customer object for go
	if err != nil { return nil, errors.New("Invalid JSON object") }
	record, err := stub.GetState(v.CustomerID) 								// If not an error then a record exists so cant create a new car with this customerId as it must be unique
	if record != nil { return nil, errors.New("Customer already exists") }
	
	_, err  = t.save_changes(stub, v)
	if err != nil { fmt.Printf("CREATE_CUSTOMER: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	bytes, err := stub.GetState("customerIDs")
	if err != nil { return nil, errors.New("Unable to get customerID") }
	var customerIDs CustomerID_Holder
	err = json.Unmarshal(bytes, &customerIDs)
	if err != nil {	return nil, errors.New("Corrupt Customer record") }
	customerIDs.Customers = append(customerIDs.Customers, customerID)
	bytes, err = json.Marshal(customerIDs)
	if err != nil { fmt.Print("Error creating Customer record") }
	err = stub.PutState("customerIDs", bytes)
	if err != nil { return nil, errors.New("Unable to put the state") }
	return nil, nil
}

//=================================================================================================================================
//	 update_name
//=================================================================================================================================
func (t *SimpleChaincode) update_name(stub shim.ChaincodeStubInterface, v Customer, new_value string) ([]byte, error) {

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

	for _, customer := range customerIDs.Customers {

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
//	 Main - main - Starts up the chaincode
//=================================================================================================================================
func main() {

	err := shim.Start(new(SimpleChaincode))
	if err != nil { fmt.Printf("Error starting Chaincode: %s", err) }
}