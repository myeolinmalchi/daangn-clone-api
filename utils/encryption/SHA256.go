package encryption

import (
    "crypto/sha256"
    "encoding/hex"
)

func EncryptSHA256(str string) string {
    hash := sha256.New()
    hash.Write([]byte(str))
    md := hash.Sum(nil)
    mdStr := hex.EncodeToString(md)
    return mdStr
}
