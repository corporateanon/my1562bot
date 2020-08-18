package apiclient

import (
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/dig"
)

type IApiClient interface {
	CreateSubscription(chatID int64, addressArID int64) error
	Geocode(lat float64, lng float64, accuracy float64) (*GeocodeResponse, error)
	FullTextSearch(query string) ([]FullTextSearchAddress, error)
	AddressStringByID(ID int64) (string, error)
	AddressByID(ID int64) (*AddressByIDResponse, error)
}

type ApiClient struct {
	client    *resty.Client
	ftsClient *resty.Client
}

type ApiClientOptions struct {
	dig.In

	Client    *resty.Client `name:"api"`
	FTSClient *resty.Client `name:"fts"`
}

func New(options ApiClientOptions) IApiClient {
	return &ApiClient{
		client:    options.Client,
		ftsClient: options.FTSClient,
	}
}

type ErrorResponse struct {
	ErrorValue string `json:"error,omitempty"`
}

func (e *ErrorResponse) Error() string {
	return e.ErrorValue
}

type CreateSubscriptionRequest struct {
	ChatID      int64
	AddressArID int64
}

func (api *ApiClient) CreateSubscription(chatID int64, addressArID int64) error {
	_, err := api.client.R().
		SetBody(CreateSubscriptionRequest{
			ChatID:      chatID,
			AddressArID: addressArID,
		}).
		Post("/subscription")
	if err != nil {
		return err
	}
	return nil
}

type GeocodeResponseAddress struct {
	ID            uint32
	Distance      float64
	AddressString string
}
type GeocodeResponse struct {
	Addresses []GeocodeResponseAddress
}

func (api *ApiClient) Geocode(
	lat float64,
	lng float64,
	accuracy float64,
) (*GeocodeResponse, error) {

	type Response struct {
		Result *GeocodeResponse
	}

	resp, err := api.client.R().
		SetResult(&Response{}).
		SetError(&ErrorResponse{}).
		SetPathParams(
			map[string]string{
				"lat":      strconv.FormatFloat(lat, 'f', -1, 64),
				"lng":      strconv.FormatFloat(lng, 'f', -1, 64),
				"accuracy": strconv.FormatFloat(accuracy, 'f', -1, 64),
			},
		).
		Get("/address-geocode/{lat}/{lng}/{accuracy}")
	if err != nil {
		return nil, err
	}
	if backendError := resp.Error(); backendError != nil {
		return nil, backendError.(*ErrorResponse)
	}
	result := resp.Result()
	return result.(*Response).Result, nil
}

type FullTextSearchAddress struct {
	ID         uint32
	Label      string
	Similarity float32
}

func (api *ApiClient) FullTextSearch(query string) ([]FullTextSearchAddress, error) {
	resp, err := api.ftsClient.R().
		SetResult([]FullTextSearchAddress{}).
		SetError(&ErrorResponse{}).
		SetBody(
			map[string]string{
				"query": query,
			},
		).
		Post("/")
	if err != nil {
		return nil, err
	}
	if backendError := resp.Error(); backendError != nil {
		return nil, backendError.(*ErrorResponse)
	}
	result := resp.Result()
	return *result.(*[]FullTextSearchAddress), nil
}

type AddressLookupResponse struct {
	Address struct {
		Address string
	}
}

func (api *ApiClient) AddressStringByID(
	ID int64,
) (string, error) {
	type Response struct {
		Result *AddressLookupResponse
	}

	resp, err := api.client.R().
		SetResult(&Response{}).
		SetError(&ErrorResponse{}).
		SetPathParams(
			map[string]string{
				"id": strconv.FormatInt(ID, 10),
			},
		).
		Get("/address-lookup/{id}")
	if err != nil {
		return "", err
	}
	if backendError := resp.Error(); backendError != nil {
		return "", backendError.(*ErrorResponse)
	}
	result := resp.Result()
	return result.(*Response).Result.Address.Address, nil
}

type AddressArCheckStatus string

const (
	AddressStatusNoWork AddressArCheckStatus = "nowork"
	AddressStatusWork                        = "work"
	AddressStatusInit                        = "init"
)

type AddressByIDResponse struct {
	ID             int64
	CheckStatus    AddressArCheckStatus
	ServiceMessage string
	Hash           string
	TakenAt        time.Time
	CheckedAt      time.Time
}

func (api *ApiClient) AddressByID(
	ID int64,
) (*AddressByIDResponse, error) {
	type Response struct {
		Result *AddressByIDResponse
	}

	resp, err := api.client.R().
		SetResult(&Response{}).
		SetError(&ErrorResponse{}).
		SetPathParams(
			map[string]string{
				"id": strconv.FormatInt(ID, 10),
			},
		).
		Get("/address/{id}")
	if err != nil {
		return nil, err
	}
	if backendError := resp.Error(); backendError != nil {
		//Normal case - address not yet exists
		if resp.StatusCode() == 404 {
			return nil, nil
		}
		return nil, backendError.(*ErrorResponse)
	}
	result := resp.Result()
	return result.(*Response).Result, nil
}
