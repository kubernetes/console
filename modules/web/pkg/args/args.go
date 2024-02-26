package args

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/dashboard/certificates/api"
	"k8s.io/klog/v2"
)

var (
	argNamespace                 = pflag.String("namespace", getEnv("POD_NAMESPACE", "kube-system"), "if non-default namespace is used encryption key will be created in the specified namespace")
	argDisableSettingsAuthorizer = pflag.Bool("disable-settings-authorizer", false, "disables settings page user authorizer so anyone can access settings page")
	argInsecurePort              = pflag.Int("insecure-port", 8000, "port to listen to for incoming HTTP requests")
	argPort                      = pflag.Int("port", 8001, "secure port to listen to for incoming HTTPS requests")
	argInsecureBindAddress       = pflag.IP("insecure-bind-address", net.IPv4(127, 0, 0, 1), "IP address on which to serve the --insecure-port, set to 127.0.0.1 for all interfaces")
	argBindAddress               = pflag.IP("bind-address", net.IPv4(0, 0, 0, 0), "IP address on which to serve the --port, set to 0.0.0.0 for all interfaces")
	argDefaultCertDir            = pflag.String("default-cert-dir", "/certs", "directory path containing files from --tls-cert-file and --tls-key-file, used also when auto-generating certificates flag is set")
	argCertFile                  = pflag.String("tls-cert-file", "", "file containing the default x509 certificate for HTTPS")
	argKeyFile                   = pflag.String("tls-key-file", "", "file containing the default x509 private key matching --tls-cert-file")
	argSystemBanner              = pflag.String("system-banner", "", "system banner message displayed in the app if non-empty, it accepts simple HTML")
	argSystemBannerSeverity      = pflag.String("system-banner-severity", "INFO", "severity of system banner, should be one of 'INFO', 'WARNING' or 'ERROR'")
	argAutoGenerateCertificates  = pflag.Bool("auto-generate-certificates", false, "enables automatic certificates generation used to serve HTTPS")
	argLocaleConfig              = pflag.String("locale-config", "./locale_conf.json", "path to file containing the locale configuration")
	argKubeconfig                = pflag.String("kubeconfig", "", "path to kubeconfig file")
)

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		value = fallback
	}
	return value
}

func init() {
	// Init klog
	fs := flag.NewFlagSet("", flag.PanicOnError)
	klog.InitFlags(fs)

	// Default log level to 1
	_ = fs.Set("v", "1")

	pflag.CommandLine.AddGoFlagSet(fs)
	pflag.Parse()
}

func Namespace() string {
	return *argNamespace
}

func DisableSettingsAuthorizer() bool {
	return *argDisableSettingsAuthorizer
}

func InsecurePort() int {
	return *argInsecurePort
}

func Port() int {
	return *argPort
}

func InsecureBindAddress() net.IP {
	return *argInsecureBindAddress
}

func BindAddress() net.IP {
	return *argBindAddress
}

func DefaultCertDir() string {
	return *argDefaultCertDir
}

func CertFile() string {
	if len(*argCertFile) == 0 && AutoGenerateCertificates() {
		return api.DashboardCertName
	}

	return *argCertFile
}

func KeyFile() string {
	if len(*argKeyFile) == 0 && AutoGenerateCertificates() {
		return api.DashboardKeyName
	}

	return *argKeyFile
}

func SystemBanner() string {
	return *argSystemBanner
}

func SystemBannerSeverity() string {
	return *argSystemBannerSeverity
}

func AutoGenerateCertificates() bool {
	return *argAutoGenerateCertificates
}

func LocaleConfig() string {
	return *argLocaleConfig
}

func InsecureAddress() string {
	return fmt.Sprintf("%s:%d", InsecureBindAddress(), InsecurePort())
}

func Address() string {
	return fmt.Sprintf("%s:%d", BindAddress(), Port())
}
func KubeconfigPath() string {
	return *argKubeconfig
}
