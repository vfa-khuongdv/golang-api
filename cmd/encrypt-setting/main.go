// Command encrypt-setting encrypts a secret setting value (e.g. mail_password)
// with SETTINGS_ENCRYPTION_KEY so it can be stored in the settings table.
//
// The value is read from stdin so it does not end up in shell history:
//
//	go run ./cmd/encrypt-setting
//	printf '%s' "$SMTP_PASSWORD" | go run ./cmd/encrypt-setting
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/vfa-khuongdv/golang-cms/internal/configs"
	"github.com/vfa-khuongdv/golang-cms/internal/shared/utils"
)

func main() {
	_ = godotenv.Load()

	key := strings.TrimSpace(configs.GetEnv("SETTINGS_ENCRYPTION_KEY", ""))

	fmt.Fprint(os.Stderr, "Value to encrypt: ")
	value, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && value == "" {
		fmt.Fprintln(os.Stderr, "\nno value read from stdin")
		os.Exit(1)
	}
	value = strings.TrimRight(value, "\r\n")

	encrypted, err := utils.EncryptSecret(key, value)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nencryption failed: %v (check SETTINGS_ENCRYPTION_KEY)\n", err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr)
	fmt.Println(encrypted)
}
