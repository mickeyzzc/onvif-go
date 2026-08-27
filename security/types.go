// Package security hosts the security/user-management domain types.
package security

import (
	"time"

	"github.com/mickeyzzc/onvif-go/v2/types"
)

// User represents a user account.
type User struct {
	Username  string
	Password  string
	UserLevel string // Administrator, Operator, User
}

// IPAddressFilter represents IP address filter.
type IPAddressFilter struct {
	Type        IPAddressFilterType
	IPv4Address []types.PrefixedIPv4Address
	IPv6Address []types.PrefixedIPv6Address
}

// IPAddressFilterType represents filter type.
type IPAddressFilterType string

// RemoteUser represents remote user configuration.
type RemoteUser struct {
	Username           string
	Password           string
	UseDerivedPassword bool
}

// Certificate represents a certificate.
type Certificate struct {
	CertificateID string
	Certificate   BinaryData
}

// BinaryData represents binary data.
type BinaryData struct {
	ContentType string
	Data        []byte
}

// CertificateStatus represents certificate status.
type CertificateStatus struct {
	CertificateID string
	Status        bool
}

// CertificateInformation represents certificate information.
type CertificateInformation struct {
	CertificateID      string
	IssuerDN           string
	SubjectDN          string
	KeyUsage           *CertificateUsage
	ExtendedKeyUsage   *CertificateUsage
	KeyLength          int
	Version            string
	SerialNum          string
	SignatureAlgorithm string
	Validity           *DateTimeRange
}

// CertificateUsage represents certificate usage.
type CertificateUsage struct {
	Critical bool
	Value    string
}

// DateTimeRange represents date/time range.
type DateTimeRange struct {
	From  time.Time
	Until time.Time
}

// AccessPolicy represents device access policy configuration.
type AccessPolicy struct {
	PolicyFile *BinaryData
}

// PasswordComplexityConfiguration represents password complexity config.
type PasswordComplexityConfiguration struct {
	MinLen                    int
	Uppercase                 int
	Number                    int
	SpecialChars              int
	BlockUsernameOccurrence   bool
	PolicyConfigurationLocked bool
}

// PasswordHistoryConfiguration represents password history config.
type PasswordHistoryConfiguration struct {
	Enabled bool
	Length  int
}

// AuthFailureWarningConfiguration represents auth failure warning config.
type AuthFailureWarningConfiguration struct {
	Enabled         bool
	MonitorPeriod   int
	MaxAuthFailures int
}

const (
	IPAddressFilterAllow IPAddressFilterType = "Allow"
	IPAddressFilterDeny  IPAddressFilterType = "Deny"
)
