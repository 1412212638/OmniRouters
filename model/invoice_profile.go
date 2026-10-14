package model

import (
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"

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
	p.Country = normalizeInvoiceCountry(p.Country)
	p.State = strings.TrimSpace(p.State)
	p.City = strings.TrimSpace(p.City)
	p.Address = strings.TrimSpace(p.Address)
	p.Postcode = strings.TrimSpace(p.Postcode)
	p.ContactName = strings.TrimSpace(p.ContactName)
}

func normalizeInvoiceCountry(country string) string {
	country = strings.TrimSpace(country)
	switch strings.ToLower(country) {
	case "中国", "中国大陆", "中华人民共和国", "china", "pr china", "people's republic of china":
		return "CN"
	case "中国香港", "香港", "hong kong":
		return "HK"
	case "中国澳门", "澳门", "macau", "macao":
		return "MO"
	case "中国台湾", "台湾", "taiwan":
		return "TW"
	case "美国", "美利坚合众国", "united states", "united states of america", "usa":
		return "US"
	case "英国", "united kingdom", "great britain", "uk":
		return "GB"
	case "日本", "japan":
		return "JP"
	case "韩国", "南韩", "south korea", "republic of korea":
		return "KR"
	case "新加坡", "singapore":
		return "SG"
	case "加拿大", "canada":
		return "CA"
	case "澳大利亚", "澳洲", "australia":
		return "AU"
	default:
		return strings.ToUpper(country)
	}
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
	if utf8.RuneCountInString(p.Email) > 320 || utf8.RuneCountInString(p.Name) > 200 || utf8.RuneCountInString(p.CompanyName) > 200 || utf8.RuneCountInString(p.TaxID) > 100 || utf8.RuneCountInString(p.Country) > 2 || utf8.RuneCountInString(p.State) > 100 || utf8.RuneCountInString(p.City) > 100 || utf8.RuneCountInString(p.Address) > 500 || utf8.RuneCountInString(p.Postcode) > 32 || utf8.RuneCountInString(p.ContactName) > 200 {
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
