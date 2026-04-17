package main

import "testing"

func TestParseSecretsdumpCredentials(t *testing.T) {
	output := `
[*] Dumping Domain Credentials (domain\uid:rid:lmhash:nthash)
CORP\Administrator:500:aad3b435b51404eeaad3b435b51404ee:8846f7eaee8fb117ad06bdd830b7586c:::
CORP\DC01$:1000:aad3b435b51404eeaad3b435b51404ee:11223344556677889900aabbccddeeff:::
CORP\Administrator:aes256-cts-hmac-sha1-96:00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff
CORP\Administrator:aes128-cts-hmac-sha1-96:00112233445566778899aabbccddeeff
`

	credentials := parseSecretsdumpCredentials(output, "CORP", "10.0.0.5")
	if len(credentials) != 4 {
		t.Fatalf("expected 4 credentials, got %d", len(credentials))
	}

	adminNTLM := credentials[0]
	if adminNTLM.Username != "Administrator" || adminNTLM.RID != "500" || adminNTLM.SecretType != "ntlm" {
		t.Fatalf("unexpected admin NTLM credential: %+v", adminNTLM)
	}
	if adminNTLM.Secret != "aad3b435b51404eeaad3b435b51404ee:8846f7eaee8fb117ad06bdd830b7586c" {
		t.Fatalf("unexpected NTLM secret: %s", adminNTLM.Secret)
	}

	machineNTLM := credentials[1]
	if machineNTLM.Username != "DC01$" || machineNTLM.Host != "DC01" || machineNTLM.RID != "1000" {
		t.Fatalf("unexpected machine credential: %+v", machineNTLM)
	}

	if credentials[2].SecretType != "kerberos-aes256" {
		t.Fatalf("expected kerberos-aes256, got %s", credentials[2].SecretType)
	}
	if credentials[3].SecretType != "kerberos-aes128" {
		t.Fatalf("expected kerberos-aes128, got %s", credentials[3].SecretType)
	}
}
