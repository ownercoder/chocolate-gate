package authenticate

import "errors"

type User struct {
	ChannelID int64
	UserName  string
}

type Config struct {
	Password string
	Lifetime int64
}

var cfg *Config
var authenticated = []User{}

func SetConfig(config *Config) {
	cfg = config
}

func Auth(ChannelID int64, UserName string, Password string) (bool, error) {
	if cfg == nil {
		return false, errors.New("auth not configured")
	}

	if IsAuthenticated(ChannelID, UserName) {
		return true, nil
	}

	if cfg.Password == Password {
		authenticate(User{
			UserName: UserName,
			ChannelID: ChannelID,
		})

		return true, nil
	}

	return false, nil
}

func authenticate(user User) {
	authenticated = append(authenticated, user)
}

func IsAuthenticated(ChannelID int64, UserName string) bool {
	for i := 0; i < len(authenticated); i++ {
		if authenticated[i].ChannelID == ChannelID && authenticated[i].UserName == UserName {
			return true
		}
	}

	return false
}
