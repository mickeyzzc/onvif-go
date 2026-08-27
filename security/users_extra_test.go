package security

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
)

const usersResponse = `<GetUsersResponse>
	<User><Username>admin</Username><UserLevel>Administrator</UserLevel></User>
	<User><Username>op</Username><UserLevel>Operator</UserLevel></User>
</GetUsersResponse>`

func TestGetUsersParses(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/device", func(action, _ string) (string, error) {
		if action != "tds:GetUsers" {
			return "", errors.New("unexpected action " + action)
		}

		return usersResponse, nil
	})

	users, err := New(caller).GetUsers(context.Background())
	if err != nil {
		t.Fatalf("GetUsers: %v", err)
	}

	if len(users) != 2 || users[0].Username != "admin" || users[1].UserLevel != "Operator" {
		t.Errorf("users = %+v", users)
	}
}

func TestCreateAndDeleteUsers(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/device", func(action, reqXML string) (string, error) {
		switch action {
		case "tds:CreateUsers":
			for _, want := range []string{"newUser", "Initialize!", "UserLevel"} {
				if !strings.Contains(reqXML, want) {
					t.Errorf("CreateUsers body misses %q: %s", want, reqXML)
				}
			}

			return `<CreateUsersResponse/>`, nil
		case "tds:DeleteUsers":
			if !strings.Contains(reqXML, "oldUser") {
				t.Errorf("DeleteUsers body misses username: %s", reqXML)
			}

			return `<DeleteUsersResponse/>`, nil
		default:
			return "", errors.New("unexpected action " + action)
		}
	})

	svc := New(caller)
	ctx := context.Background()

	if err := svc.CreateUsers(ctx, []*User{{Username: "newUser", Password: "Initialize!", UserLevel: "User"}}); err != nil {
		t.Fatalf("CreateUsers: %v", err)
	}

	if err := svc.DeleteUsers(ctx, []string{"oldUser"}); err != nil {
		t.Fatalf("DeleteUsers: %v", err)
	}
}

func TestAccessPolicyRoundTrip(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/device", func(action, reqXML string) (string, error) {
		switch action {
		case "tds:GetAccessPolicy":
			return `<GetAccessPolicyResponse>
	<AccessPolicy><DefaultAccessRules/></AccessPolicy>
</GetAccessPolicyResponse>`, nil
		case "tds:SetAccessPolicy":
			if !strings.Contains(reqXML, "AccessPolicy") {
				t.Errorf("SetAccessPolicy body: %s", reqXML)
			}

			return `<SetAccessPolicyResponse/>`, nil
		default:
			return "", errors.New("unexpected action " + action)
		}
	})

	svc := New(caller)
	ctx := context.Background()

	policy, err := svc.GetAccessPolicy(ctx)
	if err != nil {
		t.Fatalf("GetAccessPolicy: %v", err)
	}

	if err := svc.SetAccessPolicy(ctx, policy); err != nil {
		t.Fatalf("SetAccessPolicy: %v", err)
	}
}
