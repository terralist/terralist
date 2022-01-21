package settings

// sha256.Sum256([]byte("abcd"))
var EncryptSalt string = "88d4266fd4e6338d13b845fcf289579d209c897823b9217da3e161936f031589"

// sha256.Sum256([]byte("abcd"))
var CodeExchangeKey string = "88d4266fd4e6338d13b845fcf289579d209c897823b9217da3e161936f031589"

// sha256.Sum256([]byte("abcd"))
var TokenSigningSecret []byte = []byte("88d4266fd4e6338d13b845fcf289579d209c897823b9217da3e161936f031589")
