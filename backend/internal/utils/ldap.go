package utils

import (
	"crypto/tls"
	"fmt"

	"legal-riu-portal/internal/config"

	"github.com/go-ldap/ldap/v3"
)

type LDAPUserInfo struct {
	Username string
	FullName string
	Email    string
	Position string
	Division string
}

func LDAPAuthenticate(cfg config.LDAPConfig, username, password string) (*LDAPUserInfo, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("LDAP belum dikonfigurasi")
	}
	if cfg.BindDN == "" || cfg.BaseDN == "" {
		return nil, fmt.Errorf("konfigurasi LDAP belum lengkap")
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	var conn *ldap.Conn
	var err error
	if cfg.UseSSL {
		tlsCfg := &tls.Config{
			ServerName:         cfg.Host,
			InsecureSkipVerify: cfg.InsecureSkipVerify, //nolint:gosec // configured through env for development only
		}
		conn, err = ldap.DialTLS("tcp", addr, tlsCfg)
	} else {
		conn, err = ldap.Dial("tcp", addr)
	}
	if err != nil {
		return nil, fmt.Errorf("gagal terhubung ke LDAP (%s): %w", addr, err)
	}
	defer conn.Close()

	if err := conn.Bind(cfg.BindDN, cfg.BindPassword); err != nil {
		return nil, fmt.Errorf("service account bind gagal: %w", err)
	}

	filter := fmt.Sprintf(cfg.UserFilter, ldap.EscapeFilter(username))
	searchReq := ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		ldapAttributes(cfg),
		nil,
	)

	sr, err := conn.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("pencarian LDAP gagal: %w", err)
	}

	switch len(sr.Entries) {
	case 0:
		return nil, fmt.Errorf("pengguna tidak ditemukan di direktori")
	case 1:
	default:
		return nil, fmt.Errorf("ditemukan lebih dari satu entri untuk pengguna ini")
	}

	entry := sr.Entries[0]
	if err := conn.Bind(entry.DN, password); err != nil {
		return nil, fmt.Errorf("email atau password salah")
	}

	fullName := attributeValue(entry, cfg.AttrName)
	if fullName == "" {
		fullName = username
	}

	email := attributeValue(entry, cfg.AttrEmail)
	if email == "" && cfg.DefaultEmailDomain != "" {
		email = fmt.Sprintf("%s@%s", username, cfg.DefaultEmailDomain)
	}

	position := attributeValue(entry, cfg.AttrPosition)
	if position == "" {
		position = cfg.DefaultPosition
	}

	division := attributeValue(entry, cfg.AttrDivision)
	if division == "" {
		division = cfg.DefaultDivision
	}

	return &LDAPUserInfo{
		Username: username,
		FullName: fullName,
		Email:    email,
		Position: position,
		Division: division,
	}, nil
}

func ldapAttributes(cfg config.LDAPConfig) []string {
	attrs := []string{"dn"}
	for _, attr := range []string{cfg.AttrName, cfg.AttrEmail, cfg.AttrPosition, cfg.AttrDivision} {
		if attr != "" && !containsString(attrs, attr) {
			attrs = append(attrs, attr)
		}
	}
	return attrs
}

func attributeValue(entry *ldap.Entry, attr string) string {
	if attr == "" {
		return ""
	}
	return entry.GetAttributeValue(attr)
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
