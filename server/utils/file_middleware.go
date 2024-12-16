package utils

import (
	"cinema-server/domain"
	"errors"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func FileMiddleware(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		c.Abort()
		return
	}
	images := form.File["images[]"]
	if len(images) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No file uploaded"})
		c.Abort()
		return
	}
	for _, image := range images {
		if err := checkImage(image); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid file type"})
			c.Abort()
			return
		}
	}
	var uploadedImages []domain.Image
	for _, image := range images {
		if err := saveImage(c, image, &uploadedImages); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			c.Abort()
			return
		}
	}
	c.Set("file", uploadedImages)
	c.Next()
}

func DeleteFile(url string) error {
	// cloudinaryURL := os.Getenv("CLOUDINARY_URL")
	// cld, err := cloudinary.NewFromURL(cloudinaryURL)
	// if err != nil {
	// 	return err
	// }

	// publicID := strings.Split(url, "/")
	// publicID = publicID[len(publicID)-1]
	// publicID = strings.Split(publicID, ".")[0]

	// _, err = cld.Upload.Destroy(context.TODO(), uploader.DestroyParams{
	// 	PublicID: publicID,
	// })
	// if err != nil {
	// 	return err
	// }
	return nil
}

func checkImage(file *multipart.FileHeader) error {
	if file.Size > (1<<20)*5 {
		return errors.New("file size should not exceed 5MB")
	}
	extensions := strings.Split(file.Filename, ".")
	extension := extensions[len(extensions)-1]

	if extension != "jpg" && extension != "jpeg" && extension != "png" && extension != "gif" && extension != "webp" {
		return errors.New("invalid file extension")
	}
	if !strings.Contains(file.Header.Get("Content-Type"), "image") {
		return errors.New("invalid file type")
	}
	return nil
}

func saveImage(c *gin.Context, file *multipart.FileHeader, uploadedImages *[]domain.Image) error {
	// Initialize Cloudinary
	cloudinaryURL := os.Getenv("CLOUDINARY_URL")
	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		return err
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Upload the file to Cloudinary
	uploadResult, err := cld.Upload.Upload(c, src, uploader.UploadParams{
		PublicID: uuid.NewString(),
		Folder:   "uploads",
	})
	if err != nil {
		return err
	}

	*uploadedImages = append(*uploadedImages, domain.Image{
		ImageUrl: uploadResult.SecureURL,
	})
	return nil
}
