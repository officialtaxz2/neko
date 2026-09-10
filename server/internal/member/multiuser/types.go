package multiuser

import "github.com/m1k1o/neko/server/pkg/types"

type Config struct {
	AdminPassword string
	UserPassword  string
	ViewOnlyToken string
	AdminProfile  types.MemberProfile
	UserProfile   types.MemberProfile
}
