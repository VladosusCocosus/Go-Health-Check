package statusRoutes

import (
	"encoding/json"
	"fmt"
	"health-check-on-go/libs/health_check"
	"os"
	path2 "path"

	"github.com/gofiber/fiber/v2"
)

const resultPath = "./results/"

type Status struct {
}

func GetStatuses(c *fiber.Ctx) error {
	entries, err := os.ReadDir(resultPath)
	var indexFiles [][]health_check.Result

	if err != nil {
		c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	for _, entry := range entries {
		var indexFile []health_check.Result
		fmt.Println(path2.Join(resultPath, entry.Name(), "index"))
		file, err := os.ReadFile(path2.Join(resultPath, entry.Name(), "index"))

		if err != nil {
			fmt.Println(path2.Join(resultPath, entry.Name(), "index"))
			continue
		}

		err = json.Unmarshal(file, &indexFile)

		if err != nil {
			return err
		}

		indexFiles = append(indexFiles, indexFile)
	}

	return c.JSON(indexFiles)
}
