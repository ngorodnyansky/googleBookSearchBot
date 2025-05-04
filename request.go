package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type GoogleBookResponce struct {
	Kind       string  `json:"kind"`
	TotalItems int     `json:"totalItems"`
	Items      []Items `json:"items"`
}

type Items struct {
	Kind      string    `json:"kind"`
	SelfLink  string    `json:"selfLink"`
	VolmeInfo VolmeInfo `json:"volumeInfo"`
}

type VolmeInfo struct {
	Titile        string     `json:"title"`
	Subtitle      string     `json:"subtitle"`
	Authors       []string   `json:"authors"`
	Description   string     `json:"description"`
	AverageRating float32    `json:"averageRating"`
	ImageLinks    ImageLinks `json:"imageLinks"`
}

type ImageLinks struct {
	ThumbnailImageLink string `json:"thumbnail"`
	ThubnailImageBytes []byte `json:"-"`
}

func BookCoverInBytes(client http.Client, link string) []byte {
	imageResp, err := client.Get(link)
	if err != nil {
		fmt.Printf("Image download failed: %v\n", err)
		return []byte{}
	}
	defer imageResp.Body.Close()
	smallImageBytes, err := io.ReadAll(imageResp.Body)
	if err != nil {
		fmt.Println("Ошибка получения изображения")
		return []byte{}
	}
	return smallImageBytes
}

func request(message string) GoogleBookResponce {
	client := http.Client{Timeout: 3 * time.Second}
	baseUrl := "https://www.googleapis.com/books/v1/volumes"

	req, err := http.NewRequest(http.MethodGet, baseUrl, nil)
	if err != nil {
		fmt.Println("Ошибка Создания запроса!")
		return GoogleBookResponce{}
	}

	params := url.Values{}
	params.Add("q", message)

	req.URL.RawQuery = params.Encode()
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Ошибка запроса!")
		return GoogleBookResponce{}
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Ошибка чтения тела запроса!")
		return GoogleBookResponce{}
	}

	var googleResponce GoogleBookResponce
	if json.Valid(body) {
		err = json.Unmarshal(body, &googleResponce)
		if err != nil {
			fmt.Println("Ошибка чтения JSON!", err)
			return GoogleBookResponce{}
		}
	} else {
		fmt.Println("Полученный JSON файл не валиден!", err)
		return GoogleBookResponce{}
	}

	for i := 0; i < len(googleResponce.Items); i++ {
		thumbnailImageBytes := BookCoverInBytes(client, googleResponce.Items[i].VolmeInfo.ImageLinks.ThumbnailImageLink)
		googleResponce.Items[i].VolmeInfo.ImageLinks.ThubnailImageBytes = thumbnailImageBytes
	}

	return googleResponce
}
