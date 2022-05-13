package tests

import (
	"grpcFileApp/internal/services/ipHelper"
	"log"
	"testing"
)

func TestGetClienntIp(t *testing.T) {
	ip, err := ipHelper.GetClientIp()
	if err != nil {
		t.Error(err)
	}

	log.Println("内网ip为", ip)
}
