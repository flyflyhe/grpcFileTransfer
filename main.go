package main

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"grpcFileApp/internal/grpc/files"
	"io/ioutil"
	"log"
	"math"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	port = "localhost:50051"
)

type server struct {
	files.UnimplementedFileServiceServer
}

// verifyFile 校验下上传的数据包是否完整, 通过Sha256和文件数据长度两个进行判断
func (s *server) verifyFile(file []byte, hash string, length int64) bool {
	h := sha256.New()
	h.Write(file)
	myHash := fmt.Sprintf("%x", h.Sum(nil))
	log.Println("hash:", hash, " myHash:", myHash, " len:", length, " myLen:", len(file))
	return hash == myHash
}

func (s *server) Transfer(ctx context.Context, in *files.FileReq) (*files.FileRes, error) {
	if !s.verifyFile(in.File, in.Hash, in.Filelen) {
		return &files.FileRes{
			Status:  false,
			Message: "数据包哈希校验失败，请重新部署",
		}, nil
	}

	filename := "/tmp/" + in.DstDir + "/" + in.Name

	f, err := os.Create(filename)
	defer f.Close()
	if err != nil {
		return &files.FileRes{
			Status:  true,
			Message: err.Error(),
		}, nil
	}

	f.Write(in.File)

	return &files.FileRes{
		Status:  true,
		Message: "received",
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	c, err := loadTLSCredentials()
	if err != nil {
		log.Fatalf("credentials.NewServerTLSFromFile err: %v", err)
	}
	//由于要发送较大的压缩包，默认为 4M。
	//如果需要向客户端发送大文件则增加一条grpc.MaxSendMsgSize()
	s := grpc.NewServer(
		grpc.Creds(c),
		grpc.MaxRecvMsgSize(math.MaxInt64))

	//注册服务
	files.RegisterFileServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	pemClientCA, err := ioutil.ReadFile("cert/root.pem")
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}
	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair("cert/cert.pem", "cert/private.key")
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(config), nil
}
