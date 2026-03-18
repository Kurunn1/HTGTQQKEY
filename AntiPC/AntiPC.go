package AntiPC

import (
	"crypto/x509"
	"os"
	"qkey/FyneGUI"
	"strings"
)

// 这个是抓包检测，主要原理是检测是否含有证书特征
func printCertInfo(cert *x509.Certificate) {
	if strings.Contains(cert.Issuer.String(), FyneGUI.XorDecrypt([]byte{0x2A, 0x35, 0x3D, 0x3C, 0x11, 0x26, 0x24, 0x2B, 0x39, 0x75, 0x25, 0x24, 0x3B})) {
		//fmt.Println("检测到抓包", cert.Subject.String())
		os.Exit(0)
	} //HTTP DEBUGGER

}
