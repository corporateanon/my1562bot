package mock_apiclient

import "github.com/my1562/telegrambot/pkg/apiclient"

type MockApiClient struct{}

func New() apiclient.IApiClient {
	return &MockApiClient{}
}

func (api *MockApiClient) CreateSubscription(chatID int64, addressArID int64) error {
	return nil
}
func (api *MockApiClient) Geocode(
	lat float64,
	lng float64,
	accuracy float64,
) (*apiclient.GeocodeResponse, error) {
	return &apiclient.GeocodeResponse{
		Addresses: []apiclient.GeocodeResponseAddress{
			{5, 10, "221b Baker st."},
			{6, 12, "1 Lenin st."},
		},
	}, nil
}

func (api *MockApiClient) AddressStringByID(
	ID int64,
) (string, error) {
	return "221b Baker st.", nil
}

func (api *MockApiClient) AddressByID(
	ID int64,
) (*apiclient.AddressByIDResponse, error) {
	return nil, nil
}
