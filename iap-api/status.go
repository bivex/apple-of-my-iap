package iap

// Status represents an Apple IAP response status.
// See: https://developer.apple.com/library/ios/releasenotes/General/ValidateAppStoreReceipt/Chapters/ValidateRemotely.html
type Status struct {
	Code        int
	Description string
	Name        string
}

// Predefined status codes as per Apple's documentation.
var (
	ValidReceipt           = Status{Code: 0, Description: "Valid receipt.", Name: "ValidReceipt"}
	BadEnvelope           = Status{Code: 21000, Description: "The App Store could not read the JSON object you provided.", Name: "BadEnvelope"}
	BadReceipt            = Status{Code: 21002, Description: "The data in the 'receipt-data' property was malformed or missing.", Name: "BadReceipt"}
	UnauthorizedReceipt   = Status{Code: 21003, Description: "The receipt could not be authenticated.", Name: "UnauthorizedReceipt"}
	SharedSecretMismatch  = Status{Code: 21004, Description: "The shared secret you provided does not match the shared secret on file for your account.", Name: "SharedSecretMismatch"}
	ServerUnavailable     = Status{Code: 21005, Description: "The receipt server is not currently available.", Name: "ServerUnavailable"}
	SubscriptionExpired   = Status{Code: 21006, Description: "This receipt is valid but the subscription has expired.", Name: "SubscriptionExpired"}
	TestToProduction      = Status{Code: 21007, Description: "This receipt is from the test environment, but it was sent to the production environment for verification.", Name: "TestToProduction"}
	ProductionToTest      = Status{Code: 21008, Description: "This receipt is from the production environment, but it was sent to the test environment for verification.", Name: "ProductionToTest"}
)

// definedStatuses maps status codes to their defined Status objects.
var definedStatuses = map[int]Status{
	ValidReceipt.Code:          ValidReceipt,
	BadEnvelope.Code:           BadEnvelope,
	BadReceipt.Code:            BadReceipt,
	UnauthorizedReceipt.Code:   UnauthorizedReceipt,
	SharedSecretMismatch.Code:  SharedSecretMismatch,
	ServerUnavailable.Code:     ServerUnavailable,
	SubscriptionExpired.Code:   SubscriptionExpired,
	TestToProduction.Code:      TestToProduction,
	ProductionToTest.Code:      ProductionToTest,
}

// GetStatus returns the Status for a given code, or an Unknown status if the code is not defined.
func GetStatus(code int) Status {
	if s, ok := definedStatuses[code]; ok {
		return s
	}
	return Status{
		Code:        code,
		Description: "Unknown status.",
		Name:        "Unknown",
	}
}
