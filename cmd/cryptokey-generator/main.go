package main

import (
	"log"
	"os"

	"github.com/smartfor/metrics/internal/utils"
)

func main() {
	bytesPub, err := os.ReadFile("public_key.pem")
	if err != nil {
		panic(err)
	}

	bytesPriv, err := os.ReadFile("private_key.pem")
	if err != nil {
		panic(err)
	}

	str := "Hello, World!"
	v := []byte(str)
	encryptedValue, key, err := utils.EncryptWithPublicKey(v, bytesPub)
	if err != nil {
		log.Fatal(err)
	}

	out, err := utils.DecryptWithPrivateKey(encryptedValue, key, bytesPriv)
	if err != nil {
		return
	}

	log.Println("Encrypted value:\n", str == string(out))

	//// создаём шаблон сертификата
	//cert := &x509.Certificate{
	//	// указываем уникальный номер сертификата
	//	SerialNumber: big.NewInt(1658),
	//	// заполняем базовую информацию о владельце сертификата
	//	Subject: pkix.Name{
	//		Organization: []string{"Yandex.Praktikum"},
	//		Country:      []string{"RU"},
	//	},
	//	// разрешаем использование сертификата для 127.0.0.1 и ::1
	//	IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	//	// сертификат верен, начиная со времени создания
	//	NotBefore: time.Now(),
	//	// время жизни сертификата — 10 лет
	//	NotAfter:     time.Now().AddDate(10, 0, 0),
	//	SubjectKeyId: []byte{1, 2, 3, 4, 6},
	//	// устанавливаем использование ключа для цифровой подписи,
	//	// а также клиентской и серверной авторизации
	//	ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	//	KeyUsage:    x509.KeyUsageDigitalSignature,
	//}
	//
	//// создаём новый приватный RSA-ключ длиной 4096 бит
	//privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//// создаём сертификат x.509
	//certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//// кодируем сертификат и ключ в формате PEM
	//var certPEM bytes.Buffer
	//pem.Encode(&certPEM, &pem.Block{
	//	Type:  "CERTIFICATE",
	//	Bytes: certBytes,
	//})
	//
	//var privateKeyPEM bytes.Buffer
	//pem.Encode(&privateKeyPEM, &pem.Block{
	//	Type:  "RSA PRIVATE KEY",
	//	Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	//})
	//
	//// кодируем публичный ключ в формате PEM
	//var publicKeyPEM bytes.Buffer
	//pubKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//err = pem.Encode(&publicKeyPEM, &pem.Block{
	//	Type:  "PUBLIC KEY",
	//	Bytes: pubKeyBytes,
	//})
	//if err != nil {
	//	return
	//}
	//
	//// выводим сертификат и ключи
	//log.Println("Certificate:\n", certPEM.String())
	//log.Println("Private Key:\n", privateKeyPEM.String())
	//log.Println("Public Key:\n", publicKeyPEM.String()) // сохраняем сертификат и ключи в файлы
	//
	//err = os.WriteFile("certificate.pem", certPEM.Bytes(), 0644)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//err = os.WriteFile("private_key.pem", privateKeyPEM.Bytes(), 0600)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//err = os.WriteFile("public_key.pem", publicKeyPEM.Bytes(), 0644)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//log.Println("Certificate, private key, and public key have been saved to files.")
	//
	//v := []byte("Hello, World!")
	//encryptedValue, key, err := utils.EncryptWithPublicKey(v, publicKeyPEM.Bytes())
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//out, err := utils.DecryptWithPrivateKey(encryptedValue, key, privateKeyPEM.Bytes())
	//if err != nil {
	//	return
	//}
	//log.Println("Encrypted value:\n", string(out))
}
