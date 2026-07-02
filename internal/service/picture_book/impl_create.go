package picture_book

import (
	"context"

	"tool-agent/internal/common"
	"tool-agent/internal/dao"
	"tool-agent/model"
	"tool-agent/utils"

	"github.com/google/uuid"
)

type SPictureBookResp struct {
	BookId     string `json:"book_id"`
	Title      string `json:"title"`
	Icon       string `json:"icon"`
	CategoryId string `json:"category_id"`
	Status     string `json:"status"`
	Type       int    `json:"type"`
	Position   int    `json:"position"`
}

// Create 创建绘本
func (s *Service) Create(ctx context.Context, title, icon, categoryId string, bookType int, status string, position int) (common.ServiceResult, error) {
	var (
		logger = utils.SugarContext(ctx)
		result = common.NewServiceResult()
	)

	bookData := model.SPictureBook{
		BookId:     uuid.New().String(),
		Title:      title,
		Icon:       icon,
		CategoryId: categoryId,
		Type:       bookType,
		Status:     status,
		Position:   position,
	}

	if err := dao.SPictureBook.Create(&bookData); err != nil {
		logger.Errorw("PictureBookService Create dao.Create error", "error", err)
		return result, err
	}

	logger.Infow("PictureBookService Create 的值是",
		"book_id", bookData.BookId,
		"title", bookData.Title,
		"icon", bookData.Icon,
		"category_id", bookData.CategoryId,
		"type", bookData.Type,
		"status", bookData.Status,
		"position", bookData.Position,
	)

	result.Data = SPictureBookResp{
		BookId:     bookData.BookId,
		Title:      bookData.Title,
		Icon:       bookData.Icon,
		CategoryId: bookData.CategoryId,
		Status:     bookData.Status,
		Type:       bookData.Type,
		Position:   bookData.Position,
	}
	result.SetMessage("操作成功")
	return result, nil
}
