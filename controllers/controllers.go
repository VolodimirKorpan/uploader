package controllers

import (
	"context"
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/VolodimirKorpan/uploader/models"
	"github.com/VolodimirKorpan/uploader/store"
	"github.com/gin-gonic/gin"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"google.golang.org/api/option"
)

func Upload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		log.Fatalln(err)
	}

	files := form.File["upload[]"]

	for _, file := range files {
		log.Println(file.Filename)
		ext := filepath.Ext(file.Filename)
		if ext != ".jpg" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "type file not .jpg",
			})
			return
		}
		c.SaveUploadedFile(file, file.Filename)
		f, err := os.Open(file.Filename)
		if err != nil {
			log.Fatalln(err)
		}
		img, err := jpeg.Decode(f)
		if err != nil {
			log.Fatal(err)
		}

		dst := "webP/" + file.Filename[0:len(file.Filename)-len(ext)] + ".webp"
		output, err := os.Create(dst)
		if err != nil {
			log.Fatal(err)
		}
		defer output.Close()

		options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 75)
		if err != nil {
			log.Fatalln(err)
		}

		if err := webp.Encode(output, img, options); err != nil {
			log.Fatalln(err)
		}
		if err = uploadFile("test-a0e83.appspot.com", output.Name()); err != nil {
			log.Fatalln(err)
		}

		defer os.Remove(f.Name())
		defer os.Remove(output.Name())
		
		im := models.Image{
			Name: output.Name(),
		}
		store.DB.Create(&im)


		c.JSON(http.StatusOK, gin.H{
			"message": "save",
			"id": im.ID,
		})
	}

	
}

func uploadFile(bucket, object string) error {
	ctx := context.Background()
	opt := option.WithCredentialsFile("./ServiceAccountKey.json")
	client, err := storage.NewClient(ctx, opt)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	// Open local file.
	f, err := os.Open(object)
	if err != nil {
		return fmt.Errorf("os.Open: %v", err)
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	log.Println(wc.MediaLink)
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}
	return nil
}


func DownloadFile(c *gin.Context) {
	// bucket := "bucket-name"
	// object := "object-name"
	id := c.Param("id")
	var im  models.Image
	store.DB.Where("id = ?", id).First(&im)
	ctx := context.Background()
	opt := option.WithCredentialsFile("./ServiceAccountKey.json")
	client, err := storage.NewClient(ctx, opt)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	rc, err := client.Bucket("test-a0e83.appspot.com").Object(im.Name).NewReader(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer rc.Close()

	extraHeaders := map[string]string{
		"Content-Disposition": `attachment; filename="gopher.png"`,
	}

	c.DataFromReader(http.StatusOK, rc.Size(), rc.ContentType(), rc, extraHeaders)
}
