package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"

	certapi "github.com/kubernetes/dashboard/src/app/backend/cert/api"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Implements certificate Creator interface. See Creator for more information.
type ecdsaCreator struct {
	keyFile, certFile *string

	curve  elliptic.Curve
	client kubernetes.Interface
}

// GenerateKey implements certificate Creator interface. See Creator for more information.
func (self *ecdsaCreator) GenerateKey() interface{} {
	key, err := ecdsa.GenerateKey(self.curve, rand.Reader)
	if err != nil {
		log.Fatalf("[ECDSAManager] Failed to generate certificate key: %s", err)
	}

	return key
}

// GenerateCertificate implements certificate Creator interface. See Creator for more information.
func (self *ecdsaCreator) GenerateCertificate(key interface{}) []byte {
	ecdsaKey := self.getKey(key)
	pod := self.getDashboardPod()

	podDomainName := pod.Name + "." + pod.Namespace

	template := x509.Certificate{
		SerialNumber: self.generateSerialNumber(),
		Subject:      pkix.Name{CommonName: podDomainName},
		Issuer:       pkix.Name{CommonName: podDomainName},
		DNSNames:     []string{podDomainName},
	}

	if len(pod.Status.PodIP) > 0 {
		template.IPAddresses = []net.IP{net.ParseIP(pod.Status.PodIP)}
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &ecdsaKey.PublicKey, ecdsaKey)
	if err != nil {
		log.Fatalf("[ECDSAManager] Failed to create certificate: %s", err)
	}

	return certBytes
}

// StoreCertificates implements certificate Creator interface. See Creator for more information.
func (self *ecdsaCreator) StoreCertificates(path string, key interface{}, certBytes []byte) {
	ecdsaKey := self.getKey(key)
	certOut, err := os.Create(path + string(os.PathSeparator) + self.GetCertFileName())
	if err != nil {
		log.Fatalf("[ECDSAManager] Failed to open %s for writing: %s", self.GetCertFileName(), err)
	}

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	certOut.Close()

	keyOut, err := os.OpenFile(path+string(os.PathSeparator)+self.GetKeyFileName(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("[ECDSAManager] Failed to open %s for writing: %s", self.GetKeyFileName(), err)
	}

	marshaledKey, err := x509.MarshalECPrivateKey(ecdsaKey)
	if err != nil {
		log.Fatalf("[ECDSAManager] Unable to marshal %s: %v", self.GetKeyFileName(), err)
	}
	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: marshaledKey})
	keyOut.Close()
}

// GetKeyFileName implements certificate Creator interface. See Creator for more information.
func (self *ecdsaCreator) GetKeyFileName() string {
	return *self.keyFile
}

// GetCertFileName implements certificate Creator interface. See Creator for more information.
func (self *ecdsaCreator) GetCertFileName() string {
	return *self.certFile
}

func (self *ecdsaCreator) getKey(key interface{}) *ecdsa.PrivateKey {
	ecdsaKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		log.Fatal("[ECDSAManager] Key should be an instance of *ecdsa.PrivateKey")
	}

	return ecdsaKey
}

func (self *ecdsaCreator) generateSerialNumber() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("[ECDSAManager] Failed to generate serial number: %s", err)
	}

	return serialNumber
}

func (self *ecdsaCreator) getDashboardPod() *corev1.Pod {
	// These variables are populated by kubernetes downward API when using in-cluster config
	podName := os.Getenv("POD_NAME")
	podNamespace := os.Getenv("POD_NAMESPACE")
	podIP := os.Getenv("POD_IP")

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: podNamespace,
		},
		Status: corev1.PodStatus{
			PodIP: podIP,
		},
	}
}

func (self *ecdsaCreator) init() {
	if len(*self.certFile) == 0 {
		*self.certFile = certapi.DashboardCertName
	}

	if len(*self.keyFile) == 0 {
		*self.keyFile = certapi.DashboardKeyName
	}
}

// NewECDSACreator creates ECDSACreator instance.
func NewECDSACreator(keyFile, certFile *string, curve elliptic.Curve, client kubernetes.Interface) certapi.Creator {
	creator := &ecdsaCreator{
		curve:    curve,
		client:   client,
		keyFile:  keyFile,
		certFile: certFile,
	}

	creator.init()
	return creator
}
