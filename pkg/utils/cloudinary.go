package utils

import (
	"fmt"
	"drivo/internal/config"

	"github.com/cloudinary/cloudinary-go/v2"
)

func NewCloudinary() (*cloudinary.Cloudinary, error) {

	cfg, err := config.Load()

	if err != nil {
		return nil, fmt.Errorf("Cant load config: %v", err)
	}

	cloud, err := cloudinary.NewFromURL(cfg.CLOUDINARY_URL)

	if err != nil {
		return nil, fmt.Errorf("Eror creating cloudinary instance: %v", err)
	}

	return cloud, nil

}
