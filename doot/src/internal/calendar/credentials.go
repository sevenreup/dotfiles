package calendar

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var calendarScopes = []string{
	"https://www.googleapis.com/auth/calendar.readonly",
	"https://www.googleapis.com/auth/userinfo.email",
	"https://www.googleapis.com/auth/userinfo.profile",
}

// LoadConfig returns an OAuth2 config loaded from GNOME Keyring credentials.
func LoadConfig() (*oauth2.Config, error) {
	clientID, clientSecret, err := LoadCredentialValues()
	if err != nil {
		return nil, err
	}
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       calendarScopes,
		Endpoint:     google.Endpoint,
	}, nil
}
