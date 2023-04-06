package uri

import (
	"net/url"
)

func NewUri(uri string) Uri {
	url, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}

	return Uri{
		uri: uri,
		url: url,
	}
}
