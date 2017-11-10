package api

const (
	// Certificate file names that will be generated by Dashboard
	DashboardCertName = "dashboard.crt"
	DashboardKeyName  = "dashboard.key"
)

// Manager is responsible for generating and storing self-signed certificates that can be used by Dashboard
// to serve over HTTPS.
type Manager interface {
	// GenerateCertificates generates self-signed certificates.
	GenerateCertificates()
}

// Creator is responsible for preparing and generating certificates.
type Creator interface {
	// GenerateKey generates certificate key
	GenerateKey() interface{}
	// GenerateCertificate generates certificate
	GenerateCertificate(key interface{}) []byte
	// StoreCertificates saves certificates in a given path
	StoreCertificates(path string, key interface{}, certBytes []byte)
	// GetKeyFileName returns certificate key file name
	GetKeyFileName() string
	// GetCertFileName returns certificate file name
	GetCertFileName() string
}
