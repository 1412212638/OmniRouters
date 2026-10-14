package model

import (
	"errors"
	"net/mail"
	"strings"

	"github.com/QuantumNous/new-api/common"
)

// InvoiceProfile stores the user's optional, reusable billing information.
// It is deliberately kept as JSON text so SQLite, MySQL and PostgreSQL share
// the same schema and migrations.
type InvoiceProfile struct {
	Type        string `json:"type"`
	Name        string `json:"name,omitempty"`
	CompanyName string `json:"company_name,omitempty"`
	TaxID       string `json:"tax_id,omitempty"`
	Email       string `json:"email,omitempty"`
	Country     string `json:"country,omitempty"`
	State       string `json:"state,omitempty"`
	City        string `json:"city,omitempty"`
	Address     string `json:"address,omitempty"`
	Postcode    string `json:"postcode,omitempty"`
	ContactName string `json:"contact_name,omitempty"`
}

func (p *InvoiceProfile) Normalize() {
	if p == nil {
		return
	}
	p.Type = strings.ToLower(strings.TrimSpace(p.Type))
	p.Name = strings.TrimSpace(p.Name)
	p.CompanyName = strings.TrimSpace(p.CompanyName)
	p.TaxID = strings.TrimSpace(p.TaxID)
	p.Email = NormalizeEmail(p.Email)
	p.Country = strings.ToUpper(strings.TrimSpace(p.Country))
	p.State = strings.TrimSpace(p.State)
	p.City = strings.TrimSpace(p.City)
	p.Address = strings.TrimSpace(p.Address)
	p.Postcode = strings.TrimSpace(p.Postcode)
	p.ContactName = strings.TrimSpace(p.ContactName)
}

func (p *InvoiceProfile) Validate() error {
	if p == nil {
		return nil
	}
	p.Normalize()
	if p.Type != "personal" && p.Type != "company" {
		return errors.New("invoice profile type must be personal or company")
	}
	if p.Type == "personal" && p.Name == "" {
		return errors.New("personal invoice name is required")
	}
	if p.Type == "company" && p.CompanyName == "" {
		return errors.New("company invoice name is required")
	}
	if p.Email == "" {
		return errors.New("invoice email is required")
	}
	parsedEmail, err := mail.ParseAddress(p.Email)
	if err != nil || parsedEmail.Address != p.Email {
		return errors.New("invoice email is invalid")
	}
	if p.Country != "" && len(p.Country) != 2 {
		return errors.New("country must use a two-letter ISO code")
	}
	if len(p.Email) > 320 || len(p.Name) > 200 || len(p.CompanyName) > 200 || len(p.TaxID) > 100 || len(p.Country) > 2 || len(p.State) > 100 || len(p.City) > 100 || len(p.Address) > 500 || len(p.Postcode) > 32 || len(p.ContactName) > 200 {
		return errors.New("invoice profile field is too long")
	}
	return nil
}

func (user *User) GetInvoiceProfile() *InvoiceProfile {
	if user == nil || strings.TrimSpace(user.InvoiceProfile) == "" {
		return nil
	}
	profile := &InvoiceProfile{}
	if err := common.Unmarshal([]byte(user.InvoiceProfile), profile); err != nil || profile.Validate() != nil {
		return nil
	}
	return profile
}

func SetUserInvoiceProfile(userID int, profile *InvoiceProfile) error {
	if userID <= 0 {
		return errors.New("invalid user id")
	}
	var value string
	if profile != nil {
		if err := profile.Validate(); err != nil {
			return err
		}
		data, err := common.Marshal(profile)
		if err != nil {
			return err
		}
		value = string(data)
	}
	return DB.Model(&User{}).Where("id = ?", userID).Update("invoice_profile", value).Error
}
