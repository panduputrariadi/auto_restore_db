package controllers

import (
	"final-project/sekolah-beta/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func RouteCompany(app *fiber.App) {
	CompanyGroup := app.Group("/company")
	CompanyGroup.Get("/", ReadAllCompany)
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