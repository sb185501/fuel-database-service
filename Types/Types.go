package Types

type TransactionType int

const (
	TRANSACTION_TYPE_POSTPAY      TransactionType = 0
	TRANSACTION_TYPE_PREPAY       TransactionType = 1
	TRANSACTION_TYPE_ICR          TransactionType = 2
	TRANSACTION_TYPE_PREAUTH      TransactionType = 3
	TRANSACTION_TYPE_LAST_SALE    TransactionType = 4
	TRANSACTION_TYPE_OPT_EXTERNAL TransactionType = 5
)

type EventMetadata struct {
	EventId          string `json:"ev_id"`
	ResponseRequired bool   `json:"responseRequired"`
}

type HoseLimits struct {
	AuthorizationAmount  [8]int `json:"authorizationAmount"`
	MaxVolume            [8]int `json:"maxVolume"`
	MoneyLimitHoseMap    int    `json:"moneyLimitHoseMap"`
	QuantityLimitHoseMap int    `json:"quantityLimitHoseMap"`
}

type FuelSale struct {
	PumpTransactionId     int64           `json:"pumpTransactionId"`
	PosTransactionId      int             `json:"posTransactionId"`
	CreditTransactionId   int             `json:"creditTransactionId"`
	LoyaltyTransactionID  int             `json:"loyaltyTransactionID"`
	TransactionType       TransactionType `json:"transactionType"`
	Money                 int64           `json:"money"`
	Volume                int64           `json:"volume"`
	Price                 int             `json:"price"`
	OriginalPrice         int             `json:"originalPrice"`
	AuthorizationAmount   int             `json:"authorizationAmount"`
	TenderAmount          int             `json:"tenderAmount"`
	MaxVolume             int             `json:"maxVolume"`
	HoseMap               int             `json:"hoseMap"`
	TierMap               int             `json:"tierMap"`
	CurrentHose           int             `json:"currentHose"`
	CurrentTier           int             `json:"currentTier"`
	AuthorizingStation    int             `json:"authorizingStation"`
	AuthorizingCashier    int             `json:"authorizingCashier"`
	AuthorizingShift      int             `json:"authorizingShift"`
	ServiceMode           int             `json:"serviceMode"`
	AuthorizedTime        string          `json:"authorizedTime"`
	BusinessDate          string          `json:"businessDate"`
	SaleStarted           bool            `json:"saleStarted"`
	DiscountId            int             `json:"discountId"`
	DiscountReason        int             `json:"discountReason"`
	MerchTotal            int             `json:"merchTotal"`
	MerchTaxTotal         int             `json:"merchTaxTotal"`
	CardDefinitionId      int             `json:"cardDefinitionId"`
	CardNumber            string          `json:"cardNumber"`
	LoyaltyTransactionId2 int64           `json:"loyaltyTransactionId2"`
	PosTransactionUid     string          `json:"posTransactionUid"`
	FuelSaleUid           string          `json:"fuelSaleUid"`
	HoseLimits            HoseLimits      `json:"hoseLimits"`
}

type Completion struct {
	RecordNumber int    `json:"recordNumber"`
	PumpNumber   int    `json:"pumpNumber"`
	Reason       int    `json:"reason"`
	Time         string `json:"time"`
}

type SaleLockStatus struct {
	Client string `json:"client"`
	Locked bool   `json:"locked"`
}

type FlowRateData struct {
	FlowRate         int    `json:"flowRate"`
	FlowRateStatus   string `json:"flowRateStatus"`
	FuelingStartTime string `json:"fuelingStartTime"`
	FuelingEndTime   string `json:"fuelingEndTime"`
	StartTime        string `json:"startTime"`
	EndTime          string `json:"endTime"`
}

type CompletedFuelSale struct {
	Metadata     *EventMetadata `json:"metadata"`
	FuelSale     *FuelSale      `json:"sale"`
	Completion   *Completion    `json:"completion"`
	FlowRateData *FlowRateData  `json:"flowRateData"`
}

type ResponseCode int

const (
	ERR_NONE           ResponseCode = 0
	ERR_NO_RESPONSE    ResponseCode = 1
	ERR_UNSUPPORTED    ResponseCode = 2
	ERR_INVALID_SYNTAX ResponseCode = 3
	ERR_FAILED         ResponseCode = 4
	ERR_INVALID_ID     ResponseCode = 5
	ERR_WRONG_STATE    ResponseCode = 6
	ERR_NOT_READY      ResponseCode = 7
)

type FuelResponse struct {
	Cause        string       `json:"cause"`
	ResponseCode ResponseCode `json:"response"`
}

type DeleteCompletedSale struct {
	Metadata   *EventMetadata `json:"metadata"`
	SaleNumber *int64         `json:"saleNumber"`
}
