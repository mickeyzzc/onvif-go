package security

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
)

func TestSetUserEncodesCredentials(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/device", func(action, reqXML string) (string, error) {
		if action != "tds:SetUser" {
			return "", errors.New("unexpected action " + action)
		}

		for _, want := range []string{"operator1", "s3cret", "Administrator"} {
			if !strings.Contains(reqXML, want) {
				t.Errorf("SetUser body misses %q: %s", want, reqXML)
			}
		}

		return "", nil
	})

	svc := New(caller)
	if svc == nil {
		t.Fatal("New returned nil")
	}

	if err := svc.SetUser(context.Background(), &User{
		Username:  "operator1",
		Password:  "s3cret",
		UserLevel: "Administrator",
	}); err != nil {
		t.Fatalf("SetUser: %v", err)
	}
}

func TestSetUserOmitsEmptyPassword(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/device", func(_, reqXML string) (string, error) {
		if strings.Contains(reqXML, "Password") {
			t.Errorf("empty password must be omitted: %s", reqXML)
		}

		return "", nil
	})

	if err := New(caller).SetUser(context.Background(), &User{Username: "nopass", UserLevel: "User"}); err != nil {
		t.Fatalf("SetUser: %v", err)
	}
}
