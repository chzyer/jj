package internal

import "testing"

func TestUserMd5(t *testing.T) {
	md5 := GenUserPswd([]byte("hello"))
	if md5 != "0e9cdf13a8724a768fc2489adb6793c9" {
		t.Error("result not except")
	}
}
