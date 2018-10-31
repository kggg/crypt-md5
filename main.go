package main

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"
)

const alphabet = "./0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const RoundsDefault = 1000

func generate(key, salt []byte) (string, error) {
	if len(salt) == 0 {
		return "", errors.New("salt is empty")
	}
	if !bytes.HasPrefix(salt, []byte("$1$")) {
		return "", errors.New("invalid format for salt")
	}
	salts := bytes.Split(salt, []byte{'$'})

	if len(salts) < 3 {
		return "", errors.New("invalid format for salt, not have three $")
	} else {
		salt = salts[2]
	}
	if len(salt) > 8 {
		salt = salt[0:8]
	}

	A := md5.New()
	A.Write(key)
	A.Write(salt)
	A.Write(key)
	AA := A.Sum(nil)

	B := md5.New()
	B.Write(key)
	B.Write([]byte("$1$"))
	B.Write(salt)

	i := len(key)
	for ; i > 16; i -= 16 {
		B.Write(AA)
	}
	B.Write(AA[0:i])

	for i = len(key); i > 0; i >>= 1 {
		if (i & 1) == 0 {
			B.Write(key[0:1])
		} else {
			B.Write([]byte{0})
		}
	}
	Csum := B.Sum(nil)

	for i = 0; i < RoundsDefault; i++ {
		C := md5.New()

		// Add key or last result.
		if (i & 1) != 0 {
			C.Write(key)
		} else {
			C.Write(Csum)
		}
		// Add salt for numbers not divisible by 3.
		if (i % 3) != 0 {
			C.Write(salt)
		}
		// Add key for numbers not divisible by 7.
		if (i % 7) != 0 {
			C.Write(key)
		}
		// Add key or last result.
		if (i & 1) == 0 {
			C.Write(key)
		} else {
			C.Write(Csum)
		}

		Csum = C.Sum(nil)
	}

	out := make([]byte, 0, 23+len([]byte("$1$"))+len(salt))
	out = append(out, []byte("$1$")...)
	out = append(out, salt...)
	out = append(out, '$')
	out = append(out, Base64Encoding([]byte{
		Csum[12], Csum[6], Csum[0],
		Csum[13], Csum[7], Csum[1],
		Csum[14], Csum[8], Csum[2],
		Csum[15], Csum[9], Csum[3],
		Csum[5], Csum[10], Csum[4],
		Csum[11],
	})...)

	// Clean sensitive data.
	B.Reset()
	A.Reset()
	for i = 0; i < len(AA); i++ {
		AA[i] = 0
	}

	return string(out), nil
}

func Base64Encoding(src []byte) (hash []byte) {
	if len(src) == 0 {
		return []byte{} // TODO: return nil
	}

	hashSize := (len(src) * 8) / 6
	if (len(src) % 6) != 0 {
		hashSize += 1
	}
	hash = make([]byte, hashSize)

	dst := hash
	for len(src) > 0 {
		switch len(src) {
		default:
			dst[0] = alphabet[src[0]&0x3f]
			dst[1] = alphabet[((src[0]>>6)|(src[1]<<2))&0x3f]
			dst[2] = alphabet[((src[1]>>4)|(src[2]<<4))&0x3f]
			dst[3] = alphabet[(src[2]>>2)&0x3f]
			src = src[3:]
			dst = dst[4:]
		case 2:
			dst[0] = alphabet[src[0]&0x3f]
			dst[1] = alphabet[((src[0]>>6)|(src[1]<<2))&0x3f]
			dst[2] = alphabet[(src[1]>>4)&0x3f]
			src = src[2:]
			dst = dst[3:]
		case 1:
			dst[0] = alphabet[src[0]&0x3f]
			dst[1] = alphabet[(src[0]>>6)&0x3f]
			src = src[1:]
			dst = dst[2:]
		}
	}

	return
}

func RandomString(l int) string {
	bytes := []byte(alphabet)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func main() {
	salt := RandomString(9)
	salt = "$1$" + salt + "$"
	key := []byte("123456")
	k, err := generate(key, []byte(salt))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(k)
}
