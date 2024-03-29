package urlsign

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Response struct {
	SignedUrl string `json:"signed_url"`
}

type SignUrlUploadFileRequest struct {
	AlbumName  string `json:"album_name"`
	ArtistName string `json:"artist_name"`
	MusicName  string `json:"music_name"`
}

func SignUrl(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	log.Printf("SIGN URL Execute\n")
	log.Printf("%v\n", request.Body)
	body := &SignUrlUploadFileRequest{}
	if err := json.Unmarshal([]byte(request.Body), body); err != nil {
		log.Printf("Error parsing body request %v\n", err)
		return nil, err
	}
	pClient, err := getS3PreSignClient(ctx)
	if err != nil {
		return nil, err
	}

	presignUrl, err := getPresignUrl(ctx, pClient, body)
	if err != nil {
		return nil, err
	}

	response := Response{SignedUrl: presignUrl}
	b, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error parsing response %v\n", err)
		return nil, err
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(b),
	}, nil

}

func getPresignUrl(ctx context.Context, pClient *s3.PresignClient, body *SignUrlUploadFileRequest) (string, error) {
	bucketName := os.Getenv("BUCKET_NAME")
	key := formatFileName(body)
	putInput := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}
	preSignPutObject, err := pClient.PresignPutObject(ctx, putInput, s3.WithPresignExpires(time.Second*30))
	if err != nil {
		log.Printf("Error presigning the url %v\n", err)
		return "", err
	}
	return preSignPutObject.URL, nil
}

func formatFileName(body *SignUrlUploadFileRequest) string {
	re := regexp.MustCompile(`[^A-Za-z0-9]`)
	albumName := re.ReplaceAllString(body.AlbumName, "")
	artistName := re.ReplaceAllString(body.ArtistName, "")
	musicName := re.ReplaceAllString(body.MusicName, "")
	fileName := fmt.Sprintf("%s/%s/%s.txt", artistName, albumName, musicName)
	log.Printf("File name %v\n", fileName)
	return fileName
}

func getS3PreSignClient(ctx context.Context) (*s3.PresignClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Error loading config %v\n", err)
		return nil, err
	}
	s3Client := s3.NewFromConfig(cfg)
	pClient := s3.NewPresignClient(s3Client)
	return pClient, nil
}
