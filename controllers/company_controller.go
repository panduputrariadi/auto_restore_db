package controllers

import (
	"final-project/sekolah-beta/utils"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func RouteCompany(app *fiber.App) {
	CompanyGroup := app.Group("/company")
	CompanyGroup.Get("/", ReadAllCompany)
	CompanyGroup.Get("/:id", GetCompanyById)
	CompanyGroup.Get("/:id/download", DownloadCompanyHistory)
}


func ReadAllCompany(c *fiber.Ctx) error {
	companyData, err := utils.ReadAllCompany()
	if err != nil {
		logrus.Error("Error on get company list: ", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(
			map[string]any{
				"message": "Server Error",
			},
		)
	}
	return c.Status(fiber.StatusOK).JSON(
		map[string]any{
			"data":    companyData,
			"message": "Success",
		},
	)
}

func GetCompanyById(c *fiber.Ctx) error {
	companyId, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			map[string]any{
				"message": "ID not valid",
			},
		)
	}

	carData, err := utils.GetCompanyID(uint(companyId))
	if err != nil {
		if err.Error() == "record not found" {
			return c.Status(fiber.StatusNotFound).JSON(
				map[string]any{
					"message": "ID not found",
				},
			)
		}
		logrus.Error("Error on get car data: ", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(
			map[string]any{
				"message": "Server Error",
			},
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		map[string]any{
			"data":    carData,
			"message": "Success",
		},
	)
}

func DownloadCompanyHistory(c *fiber.Ctx) error {
	companyId, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			map[string]interface{}{
				"message": "ID not valid",
			},
		)
	}

	filePath, err := utils.DownloadCompanyHistoryFile(uint(companyId))
	if err != nil {
		logrus.Error("Error retrieving latest company history file: ", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(
			map[string]interface{}{
				"message": "Server Error",
			},
		)
	}
	downloadFolder := "./download"

	// Membuat folder jika belum ada
	if _, err := os.Stat(downloadFolder); os.IsNotExist(err) {
		err := os.MkdirAll(downloadFolder, 0755)
		if err != nil {
			logrus.Error("Error creating download folder: ", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(
				map[string]interface{}{
					"message": "Server Error",
				},
			)
		}
	}

	// Menentukan path lengkap file tujuan dalam folder download
	downloadPath := filepath.Join(downloadFolder, filepath.Base(filePath))

	// Membuka file asli
	sourceFile, err := os.Open(filePath)
	if err != nil {
		logrus.Error("Error opening original file: ", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(
			map[string]interface{}{
				"message": "Server Error",
			},
		)
	}
	defer sourceFile.Close()

	// Membuat file tujuan
	destinationFile, err := os.Create(downloadPath)
	if err != nil {
		logrus.Error("Error creating destination file: ", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(
			map[string]interface{}{
				"message": "Server Error",
			},
		)
	}
	defer destinationFile.Close()

	// Menyalin isi file asli ke file tujuan
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		logrus.Error("Error copying file: ", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(
			map[string]interface{}{
				"message": "Server Error",
			},
		)
	}

	return c.Download(downloadPath)
}
