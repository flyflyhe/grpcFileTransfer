package tests

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"grpcFileApp/internal/grpc/files"

	"github.com/linxlib/conv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	address = "localhost:50051"
)

func TestClient(t *testing.T) {
	tlsCredentials, err := loadTLSCredentials()
	if err != nil {
		log.Fatal("cannot load TLS credentials: ", err)
	}
	if err != nil {
		log.Fatalf("credentials.NewClientTLSFromFile err: %v", err)
	}
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(tlsCredentials))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := files.NewFileServiceClient(conn)

	// 30秒的上下文, 传输大文件适当扩大时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	bs, _ := ioutil.ReadFile("/tmp/abc.txt")
	filelen := conv.Int64(len(bs))
	h := sha256.New()
	h.Write(bs)
	fileHash := fmt.Sprintf("%x", h.Sum(nil))
	log.Println("fileHash:", fileHash)
	start := time.Now()
	r, err := client.Transfer(ctx, &files.FileReq{
		DstDir:   "ehw",
		ProjName: "dsaudg",
		Name:     "dasgf",
		ProjType: 1,
		Hash:     fileHash,
		Filelen:  filelen,
		IfReboot: false,
		File:     bs,
	})
	end := time.Now().Sub(start).Seconds()
	kb := filelen / 1024
	log.Panicln("time:", end, " file size:", kb, "KB")
	if err != nil {
		log.Fatalf("could not upload: %v", err)
	}
	log.Printf("Upload: %s", r.Message)
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed server's certificate
	pemServerCA, err := ioutil.ReadFile("../cert/root.pem")
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Load client's certificate and private key
	clientCert, err := tls.LoadX509KeyPair("../cert/client_cert.pem", "../cert/client_private.pem")
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		ServerName:   "test.com", //生成的证书通用名称 必须一致
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config), nil
}
