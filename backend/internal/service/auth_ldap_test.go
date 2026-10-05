package service

import "testing"

func TestDivisionFromDN(t *testing.T) {
	cases := map[string]string{
		"CN=Nurman Priatama Fitra,OU=IT Division,OU=Indonesiare.development,DC=indonesiare,DC=dev": "IT Division",
		"cn=Budi,ou=Legal\\, Compliance,dc=indonesiare,dc=dev":                                     "Legal, Compliance",
		"CN=Budi,CN=Users,DC=indonesiare,DC=dev":                                                   "",
		"bukan dn":                                                                                 "",
	}
	for dn, want := range cases {
		if got := divisionFromDN(dn); got != want {
			t.Errorf("divisionFromDN(%q) = %q, mau %q", dn, got, want)
		}
	}
}
