package multiuser

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"regexp"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/utils"
)

var viewOnlyTokenPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{16}$`)

func New(config Config) types.MemberProvider {
	return &MemberProviderCtx{
		config: config,
	}
}

type MemberProviderCtx struct {
	config Config
}

func (provider *MemberProviderCtx) Connect() error {
	if provider.config.ViewOnlyToken == "" {
		return nil
	}

	if !viewOnlyTokenPattern.MatchString(provider.config.ViewOnlyToken) {
		return errors.New("view-only token must contain exactly 16 URL-safe characters")
	}
	if provider.config.ViewOnlyToken == provider.config.AdminPassword ||
		provider.config.ViewOnlyToken == provider.config.UserPassword {
		return errors.New("view-only token must differ from member and admin passwords")
	}

	return nil
}

func (provider *MemberProviderCtx) Disconnect() error {
	return nil
}

func (provider *MemberProviderCtx) Authenticate(username string, password string) (string, types.MemberProfile, error) {
	// generate random token
	token, err := utils.NewUID(5)
	if err != nil {
		return "", types.MemberProfile{}, err
	}

	// id is username with token
	id := fmt.Sprintf("%s-%s", username, token)

	// The optional share token creates an ephemeral session with a fixed,
	// server-enforced profile. Compare it in constant time because it is a
	// bearer credential rather than an ordinary user-selected password.
	if provider.config.ViewOnlyToken != "" &&
		subtle.ConstantTimeCompare([]byte(provider.config.ViewOnlyToken), []byte(password)) == 1 {
		return id, types.NewViewOnlyMemberProfile(username), nil
	}

	// if logged in as administrator
	if provider.config.AdminPassword == password {
		profile := provider.config.AdminProfile
		if profile.Name == "" {
			profile.Name = username
		}
		return id, profile, nil
	}

	// if logged in as user
	if provider.config.UserPassword == password {
		profile := provider.config.UserProfile
		if profile.Name == "" {
			profile.Name = username
		}
		return id, profile, nil
	}

	return "", types.MemberProfile{}, types.ErrMemberInvalidPassword
}

func (provider *MemberProviderCtx) Insert(username string, password string, profile types.MemberProfile) (string, error) {
	return "", errors.New("new user is created on first login in multiuser mode")
}

func (provider *MemberProviderCtx) UpdateProfile(id string, profile types.MemberProfile) error {
	return nil
}

func (provider *MemberProviderCtx) UpdatePassword(id string, password string) error {
	return errors.New("password can only be modified in config while in multiuser mode")
}

func (provider *MemberProviderCtx) Select(id string) (types.MemberProfile, error) {
	return types.MemberProfile{}, errors.New("cannot select user in multiuser mode")
}

func (provider *MemberProviderCtx) SelectAll(limit int, offset int) (map[string]types.MemberProfile, error) {
	return map[string]types.MemberProfile{}, nil
}

func (provider *MemberProviderCtx) Delete(id string) error {
	return errors.New("cannot delete user in multiuser mode")
}
