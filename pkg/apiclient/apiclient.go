package apiclient

import (
	"github.com/go-resty/resty/v2"
)

type ApiClient struct {
	client *resty.Client
}

func New(client *resty.Client) *ApiClient {
	return &ApiClient{client: client}
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
