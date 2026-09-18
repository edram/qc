package qcc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// CookieInfo is safe-to-display cookie metadata; it intentionally excludes values.
type CookieInfo struct {
	Name     string
	Domain   string
	Path     string
	Expires  time.Time
	Secure   bool
	HTTPOnly bool
}

// AuthInfo identifies the QCC account associated with the client's cookies.
type AuthInfo struct {
	Nickname    string
	PhonePrefix string
	Phone       string
	Email       string
	AvatarURL   string
}

// CookieInfo returns metadata for the cookies the client uses for requests.
func (c *Client) CookieInfo(ctx context.Context) ([]CookieInfo, error) {
	cookies, err := c.cookieSource.Cookies(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]CookieInfo, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie == nil {
			continue
		}
		result = append(result, CookieInfo{
			Name:     cookie.Name,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			Expires:  cookie.Expires,
			Secure:   cookie.Secure,
			HTTPOnly: cookie.HttpOnly,
		})
	}
	return result, nil
}

// AuthInfo verifies the current cookies and returns their QCC account.
func (c *Client) AuthInfo(ctx context.Context) (AuthInfo, error) {
	response, err := c.Get(ctx, "/api/userCenter/getAuthInfo")
	if err != nil {
		return AuthInfo{}, err
	}
	defer response.Body.Close()

	var payload struct {
		Data struct {
			AvatarURL   string `json:"userAvatar"`
			Email       string `json:"userEmail"`
			Nickname    string `json:"userNickname"`
			Phone       string `json:"userPhone"`
			PhonePrefix string `json:"userPhone_prefix"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return AuthInfo{}, err
	}
	if payload.Data.Nickname == "" && payload.Data.Phone == "" && payload.Data.Email == "" {
		return AuthInfo{}, fmt.Errorf("qcc auth info: not authenticated")
	}
	return AuthInfo{
		Nickname:    payload.Data.Nickname,
		PhonePrefix: payload.Data.PhonePrefix,
		Phone:       payload.Data.Phone,
		Email:       payload.Data.Email,
		AvatarURL:   payload.Data.AvatarURL,
	}, nil
}
