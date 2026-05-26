package user

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var ErrInvalidPassword = errors.New("invalid credentials")

type argon2idParams struct {
	memoryKiB uint32
	time      uint32
	threads   uint8
	saltLen   uint32
	keyLen    uint32
}

var defaultParams = argon2idParams{
	memoryKiB: 64 * 1024,
	time:      3,
	threads:   2,
	saltLen:   16,
	keyLen:    32,
}

// Encoded form: $argon2id$v=19$m=65536,t=3,p=2$<salt_b64>$<hash_b64>
func hashPassword(password string) (string, error) {
	salt := make([]byte, defaultParams.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("rand salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		defaultParams.time,
		defaultParams.memoryKiB,
		defaultParams.threads,
		defaultParams.keyLen,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		defaultParams.memoryKiB,
		defaultParams.time,
		defaultParams.threads,
		b64Salt,
		b64Hash,
	), nil
}

func verifyPassword(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, nil
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != 19 {
		return false, nil
	}

	var mem uint32
	var iterations uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &mem, &iterations, &threads); err != nil {
		return false, nil
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, nil
	}
	wantHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, nil
	}

	gotHash := argon2.IDKey([]byte(password), salt, iterations, mem, threads, uint32(len(wantHash)))
	return subtle.ConstantTimeCompare(gotHash, wantHash) == 1, nil
}
