package security

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/mickeyzzc/onvif-go/v2/device"
	"github.com/mickeyzzc/onvif-go/v2/internal/api"
)

// Namespace aliases the device WSDL namespace for the user-management
// operations, which live in tds.
const usersNamespace = device.Namespace

// GetUsers retrieves user accounts.
func (s *Service) GetUsers(ctx context.Context) ([]*User, error) {
	type GetUsers struct {
		XMLName xml.Name `xml:"tds:GetUsers"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetUsersResponse struct {
		XMLName xml.Name `xml:"GetUsersResponse"`
		User    []struct {
			Username  string `xml:"Username"`
			UserLevel string `xml:"UserLevel"`
		} `xml:"User"`
	}

	req := GetUsers{
		Xmlns: usersNamespace,
	}

	var resp GetUsersResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetUsers failed: %w", err)
	}

	users := make([]*User, len(resp.User))
	for i, u := range resp.User {
		users[i] = &User{
			Username:  u.Username,
			UserLevel: u.UserLevel,
		}
	}

	return users, nil
}

// CreateUsers creates new user accounts.
// CreateUsers creates new user accounts.
func (s *Service) CreateUsers(ctx context.Context, users []*User) error {
	type CreateUsers struct {
		XMLName xml.Name `xml:"tds:CreateUsers"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
		User    []struct {
			Username  string `xml:"tds:Username"`
			Password  string `xml:"tds:Password"`
			UserLevel string `xml:"tds:UserLevel"`
		} `xml:"tds:User"`
	}

	req := CreateUsers{
		Xmlns: usersNamespace,
	}

	for _, user := range users {
		req.User = append(req.User, struct {
			Username  string `xml:"tds:Username"`
			Password  string `xml:"tds:Password"`
			UserLevel string `xml:"tds:UserLevel"`
		}{
			Username:  user.Username,
			Password:  user.Password,
			UserLevel: user.UserLevel,
		})
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("CreateUsers failed: %w", err)
	}

	return nil
}

// DeleteUsers deletes user accounts.
// DeleteUsers deletes user accounts.
func (s *Service) DeleteUsers(ctx context.Context, usernames []string) error {
	type DeleteUsers struct {
		XMLName  xml.Name `xml:"tds:DeleteUsers"`
		Xmlns    string   `xml:"xmlns:tds,attr"`
		Username []string `xml:"tds:Username"`
	}

	req := DeleteUsers{
		Xmlns:    usersNamespace,
		Username: usernames,
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("DeleteUsers failed: %w", err)
	}

	return nil
}

// SetUser modifies an existing user account.
// SetUser modifies an existing user account.
func (s *Service) SetUser(ctx context.Context, user *User) error {
	type SetUser struct {
		XMLName xml.Name `xml:"tds:SetUser"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
		User    struct {
			Username  string  `xml:"tds:Username"`
			Password  *string `xml:"tds:Password,omitempty"`
			UserLevel string  `xml:"tds:UserLevel"`
		} `xml:"tds:User"`
	}

	req := SetUser{
		Xmlns: usersNamespace,
	}
	req.User.Username = user.Username
	if user.Password != "" {
		req.User.Password = &user.Password
	}
	req.User.UserLevel = user.UserLevel

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("SetUser failed: %w", err)
	}

	return nil
}

// GetServices returns information about services on the device.

// GetAccessPolicy retrieves access policy information. ONVIF Specification: GetAccessPolicy operation.
func (s *Service) GetAccessPolicy(ctx context.Context) (*AccessPolicy, error) {
	type GetAccessPolicyBody struct {
		XMLName xml.Name `xml:"tds:GetAccessPolicy"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetAccessPolicyResponse struct {
		XMLName    xml.Name    `xml:"GetAccessPolicyResponse"`
		PolicyFile *BinaryData `xml:"PolicyFile"`
	}

	request := GetAccessPolicyBody{
		Xmlns: usersNamespace,
	}
	var response GetAccessPolicyResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return nil, fmt.Errorf("GetAccessPolicy failed: %w", err)
	}

	return &AccessPolicy{PolicyFile: response.PolicyFile}, nil
}

// SetAccessPolicy sets access policy information. ONVIF Specification: SetAccessPolicy operation.
// SetAccessPolicy sets access policy information. ONVIF Specification: SetAccessPolicy operation.
func (s *Service) SetAccessPolicy(ctx context.Context, policy *AccessPolicy) error {
	type SetAccessPolicyBody struct {
		XMLName    xml.Name    `xml:"tds:SetAccessPolicy"`
		Xmlns      string      `xml:"xmlns:tds,attr"`
		PolicyFile *BinaryData `xml:"tds:PolicyFile"`
	}

	type SetAccessPolicyResponse struct {
		XMLName xml.Name `xml:"SetAccessPolicyResponse"`
	}

	request := SetAccessPolicyBody{
		Xmlns:      usersNamespace,
		PolicyFile: policy.PolicyFile,
	}
	var response SetAccessPolicyResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", request, &response); err != nil {
		return fmt.Errorf("SetAccessPolicy failed: %w", err)
	}

	return nil
}

// GetWsdlURL retrieves the WSDL URL (deprecated). ONVIF Specification: GetWsdlUrl operation.
