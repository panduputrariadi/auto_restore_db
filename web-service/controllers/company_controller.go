package controllers

import (
	"final-project/sekolah-beta/middleware"
	"final-project/sekolah-beta/utils"
	"io"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func RouteCompany(app *fiber.App) {
	CompanyGroup := app.Group("/company", middleware.CheckClient)
	CompanyGroup.Get("/", middleware.CheckClient, ReadAllCompany)
	CompanyGroup.Get("/search", middleware.CheckClient, GetCompanyById)
	CompanyGroup.Get("/download", middleware.CheckClient, DownloadCompanyHistory)
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
	companyName := c.Query("company_name")

	if companyName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			map[string]interface{}{
				"message": "Company name is required in query parameter",
			},
		)
	}

	companyData, err := utils.GetCompanyID(companyName)
	if err != nil {
		if err.Error() == "record not found" {
			return c.Status(fiber.StatusNotFound).JSON(
				map[string]interface{}{
					"message": "Company not found",
				},
			)
		}
		logrus.Error("Error on get company data: ", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(
			map[string]interface{}{
				"message": "Server Error",
			},
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		map[string]interface{}{
			"data":    companyData,
			"message": "Success",
		},
	)
}

func DownloadCompanyHistory(c *fiber.Ctx) error {
	companyName := c.Query("company_name")

	if companyName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			map[string]interface{}{
				"message": "Company name is required in query parameter",
			},
		)
	}

	filePath, err := utils.DownloadCompanyHistoryFile(companyName)
	if err != nil {
		logrus.Error("Error retrieving latest company history file: ", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]interface{}{
			"message": "Server Error",
		})
	}

	sourceFile, err := os.Open(filePath)
	if err != nil {
		logrus.Error("Error opening original file: ", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]interface{}{
			"message": "Server Error",
		})
	}
	defer sourceFile.Close()

	c.Set(fiber.HeaderContentDisposition, "attachment; filename="+filepath.Base(filePath))

	if _, err := io.Copy(c, sourceFile); err != nil {
		logrus.Error("Error copying file to response: ", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]interface{}{
			"message": "Server Error",
		})
	}

	return nil
}
