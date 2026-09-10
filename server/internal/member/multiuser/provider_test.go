package multiuser

import (
	"strings"
	"testing"

	"github.com/m1k1o/neko/server/pkg/types"
)

const validViewOnlyToken = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestViewOnlyTokenAuthentication(t *testing.T) {
	provider := &MemberProviderCtx{config: Config{
		AdminPassword: "admin-password",
		UserPassword:  "member-password",
		ViewOnlyToken: validViewOnlyToken,
	}}

	if err := provider.Connect(); err != nil {
		t.Fatalf("valid configuration rejected: %v", err)
	}

	id, profile, err := provider.Authenticate("Guest", validViewOnlyToken)
	if err != nil {
		t.Fatalf("view-only authentication failed: %v", err)
	}
	if !strings.HasPrefix(id, "Guest-") {
		t.Fatalf("session id %q does not preserve display name prefix", id)
	}
	if profile.Name != "Guest" || !profile.IsViewOnly || !profile.CanWatch {
		t.Fatalf("unexpected view-only profile: %+v", profile)
	}
	if profile.IsAdmin || profile.CanHost || profile.CanShareMedia || profile.CanAccessClipboard {
		t.Fatalf("view-only token granted an interactive capability: %+v", profile)
	}
}

func TestViewOnlyTokenDisabled(t *testing.T) {
	provider := &MemberProviderCtx{config: Config{
		AdminPassword: "admin-password",
		UserPassword:  "member-password",
	}}

	_, _, err := provider.Authenticate("Guest", validViewOnlyToken)
	if err != types.ErrMemberInvalidPassword {
		t.Fatalf("disabled view-only token error = %v, want %v", err, types.ErrMemberInvalidPassword)
	}
}

func TestViewOnlyTokenRejectsDifferentValidShape(t *testing.T) {
	provider := &MemberProviderCtx{config: Config{
		AdminPassword: "admin-password",
		UserPassword:  "member-password",
		ViewOnlyToken: validViewOnlyToken,
	}}

	wrongToken := "1123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	_, _, err := provider.Authenticate("Guest", wrongToken)
	if err != types.ErrMemberInvalidPassword {
		t.Fatalf("wrong view-only token error = %v, want %v", err, types.ErrMemberInvalidPassword)
	}
}

func TestViewOnlyTokenConfigurationValidation(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "wrong shape",
			config: Config{
				AdminPassword: "admin-password",
				UserPassword:  "member-password",
				ViewOnlyToken: "short",
			},
		},
		{
			name: "admin collision",
			config: Config{
				AdminPassword: validViewOnlyToken,
				UserPassword:  "member-password",
				ViewOnlyToken: validViewOnlyToken,
			},
		},
		{
			name: "member collision",
			config: Config{
				AdminPassword: "admin-password",
				UserPassword:  validViewOnlyToken,
				ViewOnlyToken: validViewOnlyToken,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := &MemberProviderCtx{config: test.config}
			if err := provider.Connect(); err == nil {
				t.Fatal("invalid view-only token configuration was accepted")
			}
		})
	}
}
