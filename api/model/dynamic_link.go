package model

type ExchangeShortLinkRequest struct {
	RequestedLink string `json:"requestedLink"`
}

type LongLinkResponseModel struct {
	LongLink string `json:"longLink"`
}
