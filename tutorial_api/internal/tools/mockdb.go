package tools

type mockDB struct{}

var mockLoginDetails = map[string]LoginDetails{
	"christian": {
		AuthToken: "123ABC",
		Username:  "christian",
	},
}

var mockCoinDetails = map[string]CoinDetails{
	"christian": {
		Coins:    100,
		Username: "christian",
	},
}

func (d *mockDB) GetUserLoginDetails(username string) *LoginDetails {
	var clientData = LoginDetails{}
	clientData, ok := mockLoginDetails[username]

	if !ok {
		return nil
	}

	return &clientData
}

func (d *mockDB) GetUserCoins(username string) *CoinDetails {
	var clientData = CoinDetails{}
	clientData, ok := mockCoinDetails[username]
	if !ok {
		return nil
	}

	return &clientData
}

func (d *mockDB) SetupDatabase() error {
	return nil
}
