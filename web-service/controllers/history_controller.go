package controllers

import (
	"final-project/sekolah-beta/model"
	"final-project/sekolah-beta/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func RouteHistory(app *fiber.App) {
	HistoryGroup := app.Group("/history")
	HistoryGroup.Post("/", InsertHistoryData)
}

func InsertHistoryData(c *fiber.Ctx) error {
	// Mendapatkan nilai dari form data
	databaseNameStr := c.FormValue("database_name")

	// Memeriksa apakah nilai database_name tidak kosong
	if databaseNameStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"message": "database_name is required",
		})
	}

	// Mengonversi nilai databaseNameStr ke dalam tipe int
	databaseName, err := strconv.Atoi(databaseNameStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"message": "invalid database_name format",
		})
	}

	// Upload file
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"message": "failed to upload file",
		})
	}

	// Memeriksa apakah file memiliki ekstensi yang diperbolehkan (zip atau rar)
	if !isValidFileExtension(fileHeader) {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"message": "invalid file extension, only zip and rar are allowed",
		})
	}

	// Membuat nama file baru dengan format "mysql_tanggaluuid"
	newFileName := "mysql_" + time.Now().Format("2006-01-02") + "_" + uuid.New().String()

	// Salin file ke dalam direktori upload dengan nama baru
	uploadedFilePath, err := saveUploadedFile(fileHeader, newFileName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]interface{}{
			"message": "failed to save uploaded file",
		})
	}

	// Memasukkan data ke dalam database
	historyData, errCreateHistory := utils.InsertHistoryData(model.History{
		DatabaseName: databaseName,
		File:         uploadedFilePath,
	})
	if errCreateHistory != nil {
		logrus.Printf("Terjadi error : %s\n", errCreateHistory.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]interface{}{
			"message": "server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"data":    historyData,
		"message": "success create data",
	})
}

// isValidFileExtension memeriksa apakah file memiliki ekstensi yang diperbolehkan (zip atau rar)
func isValidFileExtension(fileHeader *multipart.FileHeader) bool {
	// Mendapatkan ekstensi file
	ext := filepath.Ext(fileHeader.Filename)
	ext = strings.ToLower(ext)

	// Memeriksa apakah ekstensi file adalah zip atau rar
	return ext == ".zip" || ext == ".rar"
}

// saveUploadedFile menyimpan file yang di-upload ke dalam direktori upload dan mengembalikan path file yang disimpan
func saveUploadedFile(fileHeader *multipart.FileHeader, newFileName string) (string, error) {
	// Membuat direktori upload jika belum ada
	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		err := os.Mkdir(uploadDir, 0755)
		if err != nil {
			return "", err
		}
	}

	// Membuka file yang di-upload
	fileSrc, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer fileSrc.Close()

	// Membuat path untuk menyimpan file
	filePath := filepath.Join(uploadDir, newFileName+filepath.Ext(fileHeader.Filename))

	// Membuka file tujuan untuk menulis
	fileDst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer fileDst.Close()

	// Menyalin isi file ke file tujuan
	_, err = io.Copy(fileDst, fileSrc)
	if err != nil {
		return "", err
	}

	return filePath, nil
}
