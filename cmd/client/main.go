package main

import (
	"bufio"
	"fmt"
	"github.com/go-resty/resty/v2"
	"os"
	"strings"
)

func main() {
	endpoint := "http://localhost:8080/"
	fmt.Println("Введите длинный URL")
	reader := bufio.NewReader(os.Stdin)
	long, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения: %v\n", err)
		return
	}
	long = strings.TrimSuffix(long, "\n")
	client := resty.New()

	response, err := client.R().
		SetBody(strings.NewReader(long)).
		SetHeader("Content-Type", "text/plain").
		Post(endpoint)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка отправки: %v\n", err)
		return
	}
	fmt.Println("Статус-код ", response.Status())
	fmt.Println("Тело ответа ", response)
}
