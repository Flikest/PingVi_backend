package servicehttp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *ServiceS3) UploadToS3(ctx *gin.Context) {
	reader, err := ctx.Request.MultipartReader()
	if err != nil {
		s.Log.Error("error with creating http reader: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	part, err := reader.NextPart()
	if err != nil {
		s.Log.Error("error with getting part from reader: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if part.FormName() != "file" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Ожидается поле 'file'"})
		return
	}

}
