package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/vancho-go/url-shortener/internal/app/config"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	config, err := config.ParseClient()
	if err != nil {
		panic(errors.New("error parsing client config"))
	}
	// контейнер данных для запроса
	data := url.Values{}
	// приглашение в консоли
	fmt.Println("Введите длинный URL")
	// открываем потоковое чтение из консоли
	reader := bufio.NewReader(os.Stdin)
	// читаем строку из консоли
	long, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	long = strings.TrimSuffix(long, "\n")
	// заполняем контейнер данными
	data.Set("url", long)
	// добавляем HTTP-клиент
	client := &http.Client{}
	// пишем запрос
	// запрос методом POST должен, помимо заголовков, содержать тело
	// тело должно быть источником потокового чтения io.Reader
	request, err := http.NewRequest(http.MethodPost, config.ClientHost, strings.NewReader(data.Encode()))
	if err != nil {
		panic(err)
	}
	// в заголовках запроса указываем кодировку
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// отправляем запрос и получаем ответ
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	// выводим код ответа
	fmt.Println("Статус-код ", response.Status)
	defer response.Body.Close()
	// читаем поток из тела ответа
	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(errors.New("error reading body"))
	}
	// и печатаем его
	fmt.Println(string(body))
}
