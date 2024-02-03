package utils

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var ErrorWrongURLFlag = errors.New("wrong url flag")
var ErrorWrongOrderNumber = errors.New("wrong order number")

func GetValidURL(urlString string) (string, error) {

	if !strings.HasPrefix(urlString, "http") {
		urlString = "http://" + urlString
	}
	u, err := url.Parse(urlString)
	if err != nil {
		return "", ErrorWrongURLFlag
	}

	if !u.IsAbs() || u.Port() == "" || u.Hostname() == "" {
		return "", ErrorWrongURLFlag
	}
	return u.Host, nil

}

func OnDialErr(err error) bool {
	var oe *net.OpError
	if errors.As(err, &oe) {
		return oe.Op == "dial"
	}
	return false
}

func IsValidOrderNumber(orderNumber string) (bool, error) {
	// Проверка наличия только цифр в номере карты
	for _, char := range orderNumber {
		if char < '0' || char > '9' {
			return false, ErrorWrongOrderNumber
		}
	}

	sum := 0
	double := false

	// Итерация по цифрам номера карты справа налево
	for i := len(orderNumber) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(orderNumber[i]))
		if err != nil {
			return false, err
		}

		// Удваиваем каждую вторую цифру, начиная с предпоследней
		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		// Суммируем цифры
		sum += digit
		double = !double
	}

	// Карта валидна, если сумма кратна 10
	return sum%10 == 0, nil
}

func FindProjectRoot() (string, error) {

	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(currentDir, "go.mod")); err == nil {
			return currentDir, nil
		}
		parentDir := filepath.Dir(currentDir)

		if parentDir == currentDir {
			break
		}
		currentDir = parentDir
	}

	return "", fmt.Errorf("go.mod not found")
}
