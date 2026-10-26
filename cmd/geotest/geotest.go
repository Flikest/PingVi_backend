package main

import (
	"fmt"

	tokens "github.com/Flikest/PingVi_backend/pkg/tokens"
	"github.com/google/uuid"
)

func main() {

	uid := uuid.New()

	p, err := tokens.CreateAccessToken(uid, []byte("OHFoiu43ouf*H$(R#O()IhKKBNibnofu8iynwqiokj90irffg9h98f439t3gh)"))
	if err != nil {
		panic(err)
	}

	fmt.Println(p)

	r, err := tokens.Verify(p, []byte("OHFoiu43ouf*H$(R#O()IhKKBNibnofu8iynwqiokj90irffg9h98f439t3gh)"))
	if err != nil {
		panic(err)
	}

	fmt.Println(r)

}
