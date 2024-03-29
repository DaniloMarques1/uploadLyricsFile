package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
)

type fileInfo struct {
	ArtistName string `json:"artist_name"`
	AlbumName  string `json:"album_name"`
	MusicName  string `json:"music_name"`
}

func NewFileInfo(artistName, albumName, musicName string) *fileInfo {
	return &fileInfo{ArtistName: artistName, AlbumName: albumName, MusicName: musicName}
}

func main() {
	fileInfo := readFileInfo()
	file, err := readFile()
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	b, err := json.Marshal(fileInfo)
	if err != nil {
		log.Fatal(err)
	}
	buf := bytes.NewBuffer(b)
	fs := NewFileService("https://rj9lagufy3.execute-api.localhost.localstack.cloud:4566/dev/upload")
	signedUrl, err := fs.RequestSignedUrl(buf)
	if err != nil {
		log.Fatal(err)
	}
	if err := fs.UploadFile(signedUrl, file); err != nil {
		log.Fatal(err)
	}
	fmt.Println("File uploaded successfully")
}

func readFile() (*os.File, error) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Provide the absolute file path: ")
	var filePath string
	if scanner.Scan() {
		filePath = string(scanner.Bytes())
	}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func readFileInfo() *fileInfo {
	var artistName, albumName, musicName string
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Artist name: ")
	if scanner.Scan() {
		artistName = string(scanner.Bytes())
	}
	fmt.Print("Album name: ")
	if scanner.Scan() {
		albumName = string(scanner.Bytes())
	}
	fmt.Print("Music name: ")
	if scanner.Scan() {
		musicName = string(scanner.Bytes())
	}
	return NewFileInfo(artistName, albumName, musicName)
}

type UrlSignResponse struct {
	SignedUrl string `json:"signed_url"`
}

type fileService struct {
	httpClient *http.Client
	url        string
}

func NewFileService(url string) *fileService {
	return &fileService{httpClient: http.DefaultClient, url: url}
}

func (fs *fileService) RequestSignedUrl(request *bytes.Buffer) (string, error) {
	req, err := http.NewRequest(http.MethodPut, fs.url, request)
	if err != nil {
		return "", err
	}
	resp, err := fs.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("Status Code: %v\n", resp.StatusCode)
		return "", errors.New("Something went wrong with the request")
	}
	urlSignResponse := &UrlSignResponse{}
	if err := json.NewDecoder(resp.Body).Decode(urlSignResponse); err != nil {
		return "", err
	}
	return urlSignResponse.SignedUrl, nil
}

func (fs *fileService) UploadFile(signedUrl string, file *os.File) error {
	req, err := http.NewRequest(http.MethodPut, signedUrl, file)
	if err != nil {
		return err
	}
	resp, err := fs.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("Status Code: %v\n", resp.StatusCode)
		return errors.New("Something went wrong with the request")
	}
	return nil
}
