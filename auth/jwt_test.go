package auth

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestGenerateJWT(t *testing.T) {
	user := User{Id: 435}

	jwt := generateJWT(user)
	split := strings.Split(jwt, ".")

	if len(split) != 3 {
		t.Fatal("jwt should have the 3 parts: header, payload and sign")
	}

	p, err := base64.RawURLEncoding.DecodeString(split[1])
	check(err, t)

	var payloadJson jwtPayload
	err = json.Unmarshal(p, &payloadJson)
	check(err, t)

	if payloadJson.UserId != user.Id {
		t.Fatalf("userId should be %d but is %d", payloadJson.UserId, user.Id)
	}
}

func check(err error, t *testing.T) {
	if err != nil {
		t.Fatal(err)
	}
}
