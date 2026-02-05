package clerk

// EventType is a type of webhook event.
type EventType string

const (
	UserCreated EventType = "user.created"
	UserUpdated EventType = "user.updated"
	UserDeleted EventType = "user.deleted"
)

type Event interface {
	GetType() EventType
	GetTimestamp() int64
	GetObject() string
}

type BaseEvent struct {
	Object          string          `json:"object"`
	Type            EventType       `json:"type"`
	Timestamp       int64           `json:"timestamp"`
	EventAttributes EventAttributes `json:"data"`
}

func (e BaseEvent) GetType() EventType {
	return e.Type
}

func (e BaseEvent) GetTimestamp() int64 {
	return e.Timestamp
}

func (e BaseEvent) GetObject() string {
	return e.Object
}

type EventAttributes struct {
	HTTPRequest HTTPRequest `json:"request"`
}

type HTTPRequest struct {
	ClientIP  string `json:"client_ip"`
	UserAgent string `json:"user_agent"`
}

type EmailAddress struct {
	ID           string       `json:"id"`
	EmailAddress string       `json:"email_address"`
	Reserved     bool         `json:"reserved"`
	Verification Verification `json:"verification"`
	LinkedTo     []any        `json:"linked_to"`
	Object       string       `json:"object"`
}

type Verification struct {
	Status   string `json:"status"`
	Strategy string `json:"strategy"`
	Attempts *int   `json:"attempts"`
	ExpireAt *int64 `json:"expire_at"`
}

// User represents a full Clerk user object
type ClerkUser struct {
	ID                            string         `json:"id"`
	Object                        string         `json:"object"`
	ExternalID                    *string        `json:"external_id"`
	PrimaryEmailAddressID         *string        `json:"primary_email_address_id"`
	PrimaryPhoneNumberID          *string        `json:"primary_phone_number_id"`
	PrimaryWeb3WalletID           *string        `json:"primary_web3_wallet_id"`
	Username                      *string        `json:"username"`
	FirstName                     *string        `json:"first_name"`
	LastName                      *string        `json:"last_name"`
	ProfileImageURL               string         `json:"profile_image_url"`
	ImageURL                      string         `json:"image_url"`
	HasImage                      bool           `json:"has_image"`
	PublicMetadata                map[string]any `json:"public_metadata"`
	PrivateMetadata               map[string]any `json:"private_metadata"`
	UnsafeMetadata                map[string]any `json:"unsafe_metadata"`
	EmailAddresses                []EmailAddress `json:"email_addresses"`
	PhoneNumbers                  []any          `json:"phone_numbers"`
	Web3Wallets                   []any          `json:"web3_wallets"`
	ExternalAccounts              []any          `json:"external_accounts"`
	SamlAccounts                  []any          `json:"saml_accounts"`
	Passkeys                      []any          `json:"passkeys"`
	EnterpriseAccounts            []any          `json:"enterprise_accounts"`
	PasswordEnabled               bool           `json:"password_enabled"`
	TwoFactorEnabled              bool           `json:"two_factor_enabled"`
	TOTPEnabled                   bool           `json:"totp_enabled"`
	BackupCodeEnabled             bool           `json:"backup_code_enabled"`
	MFAEnabledAt                  *int64         `json:"mfa_enabled_at"`
	MFADisabledAt                 *int64         `json:"mfa_disabled_at"`
	CreateOrganizationEnabled     bool           `json:"create_organization_enabled"`
	CreateOrganizationsLimit      *int           `json:"create_organizations_limit"`
	DeleteSelfEnabled             bool           `json:"delete_self_enabled"`
	Banned                        bool           `json:"banned"`
	Locked                        bool           `json:"locked"`
	LockoutExpiresInSeconds       *int           `json:"lockout_expires_in_seconds"`
	VerificationAttemptsRemaining *int           `json:"verification_attempts_remaining"`
	CreatedAt                     int64          `json:"created_at"`
	UpdatedAt                     int64          `json:"updated_at"`
	LastSignInAt                  *int64         `json:"last_sign_in_at"`
	LastActiveAt                  *int64         `json:"last_active_at"`
	LegalAcceptedAt               *int64         `json:"legal_accepted_at"`
}

// Clerk dashboard is setup to require an email however it is possible to not set a primary email address
// so we will return the first email address if the primary is not set
func (cl *ClerkUser) GetPrimaryOrFirstEmailAddress() *EmailAddress {
	if len(cl.EmailAddresses) == 0 {
		return nil
	}

	if cl.PrimaryEmailAddressID != nil {
		for _, emailAddress := range cl.EmailAddresses {
			if emailAddress.ID == *cl.PrimaryEmailAddressID {
				return &emailAddress
			}
		}
	}

	return &cl.EmailAddresses[0]
}

type DeletedUser struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

type UserCreatedEvent struct {
	BaseEvent
	Data ClerkUser `json:"data"`
}

type UserUpdateEvent struct {
	BaseEvent
	Data ClerkUser `json:"data"`
}

type UserDeletedEvent struct {
	BaseEvent
	Data DeletedUser `json:"data"`
}
