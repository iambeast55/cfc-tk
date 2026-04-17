package main

// Team represents a team entity
type Team struct {
	SubnetId int      `json:"subnetId"`
	Name     string   `json:"name"`
	Targets  []string `json:"targets"`
}

// CreateTeamRequest is the request payload for creating a team
type CreateTeamRequest struct {
	Name     string `json:"name"`
	SubnetId int    `json:"subnetId"`
}

// UpdateTeamRequest is the request payload for updating a team
type UpdateTeamRequest struct {
	Name     string `json:"name"`
	SubnetId int    `json:"subnetId"`
}

type AddTargetToTeam struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
	OS   string `json:"os"`
}

type RemoveTargetFromTeam struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}

type Credential struct {
	ID         int    `json:"id"`
	TeamName   string `json:"teamName"`
	OS         string `json:"os"`
	Username   string `json:"username"`
	SecretType string `json:"secretType"`
	Secret     string `json:"secret"`
	RID        string `json:"rid"`
	Domain     string `json:"domain"`
	Host       string `json:"host"`
	IP         string `json:"ip"`
	CreatedAt  string `json:"createdAt"`
}

type CreateCredentialRequest struct {
	OS         string `json:"os"`
	Username   string `json:"username"`
	SecretType string `json:"secretType"`
	Secret     string `json:"secret"`
	RID        string `json:"rid"`
	Domain     string `json:"domain"`
	Host       string `json:"host"`
	IP         string `json:"ip"`
}

type Domain struct {
	ID        int    `json:"id"`
	TeamName  string `json:"teamName"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

type CreateDomainRequest struct {
	Name string `json:"name"`
}

type Target struct {
	ID         int    `json:"id"`
	TeamName   string `json:"teamName"`
	DomainID   *int   `json:"domainId"`
	DomainName string `json:"domainName"`
	Hostname   string `json:"hostname"`
	IP         string `json:"ip"`
	OS         string `json:"os"`
	CreatedAt  string `json:"createdAt"`
}

type CreateTargetRequest struct {
	Hostname   string `json:"hostname"`
	IP         string `json:"ip"`
	OS         string `json:"os"`
	DomainID   *int   `json:"domainId"`
	DomainName string `json:"domainName"`
}

type KerberosCache struct {
	ID        int    `json:"id"`
	TeamName  string `json:"teamName"`
	Domain    string `json:"domain"`
	Username  string `json:"username"`
	Method    string `json:"method"`
	CachePath string `json:"cachePath"`
	KDCHost   string `json:"kdcHost"`
	DomainSID string `json:"domainSid"`
	UserID    string `json:"userId"`
	Groups    string `json:"groups"`
	ExtraSID  string `json:"extraSid"`
	Duration  string `json:"duration"`
	ExpiresAt string `json:"expiresAt"`
	Notes     string `json:"notes"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type CreateKerberosCacheRequest struct {
	Domain    string `json:"domain"`
	Username  string `json:"username"`
	Method    string `json:"method"`
	CachePath string `json:"cachePath"`
	KDCHost   string `json:"kdcHost"`
	DomainSID string `json:"domainSid"`
	UserID    string `json:"userId"`
	Groups    string `json:"groups"`
	ExtraSID  string `json:"extraSid"`
	Duration  string `json:"duration"`
	ExpiresAt string `json:"expiresAt"`
	Notes     string `json:"notes"`
}

type RunKerberosTicketRequest struct {
	Domain         string `json:"domain"`
	Username       string `json:"username"`
	Method         string `json:"method"`
	ToolCommand    string `json:"toolCommand"`
	TicketAuthMode string `json:"ticketAuthMode"`
	Password       string `json:"password"`
	LMHash         string `json:"lmHash"`
	NTHash         string `json:"ntHash"`
	AESKey         string `json:"aesKey"`
	KrbTGTAESKey   string `json:"krbtgtAesKey"`
	KDCHost        string `json:"kdcHost"`
	DomainSID      string `json:"domainSid"`
	UserID         string `json:"userId"`
	Groups         string `json:"groups"`
	ExtraSID       string `json:"extraSid"`
	Duration       string `json:"duration"`
	ExpiresAt      string `json:"expiresAt"`
	Notes          string `json:"notes"`
}

type RunKerberosTicketResponse struct {
	Cache   *KerberosCache `json:"cache"`
	Command []string       `json:"command"`
	Output  string         `json:"output"`
}

type RunSecretsdumpRequest struct {
	ToolCommand      string `json:"toolCommand"`
	Target           string `json:"target"`
	Domain           string `json:"domain"`
	Username         string `json:"username"`
	AuthMode         string `json:"authMode"`
	Password         string `json:"password"`
	LMHash           string `json:"lmHash"`
	NTHash           string `json:"ntHash"`
	AESKey           string `json:"aesKey"`
	KDCHost          string `json:"kdcHost"`
	UseKerberosCache bool   `json:"useKerberosCache"`
	CachePath        string `json:"cachePath"`
	JustDC           bool   `json:"justDc"`
	UseVSS           bool   `json:"useVss"`
}

type RunSecretsdumpResponse struct {
	Command     []string     `json:"command"`
	Output      string       `json:"output"`
	Credentials []Credential `json:"credentials"`
}
