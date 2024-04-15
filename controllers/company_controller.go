package controllers

import (
	"final-project/sekolah-beta/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func RouteCompany(app *fiber.App) {
	CompanyGroup := app.Group("/company")
	CompanyGroup.Get("/", ReadAllCompany)
	CompanyGroup.Get("/:id", GetCompanyById)
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