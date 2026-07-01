package picture_book

import (
	"context"

	"tool-agent/internal/common"
	"tool-agent/internal/dao"
	"tool-agent/model"
	"tool-agent/utils"

	"github.com/google/uuid"
)

// Create 创建绘本
func (s *Service) Create(ctx context.Context, title, icon, categoryId string, bookType int, status string, position int) (common.ServiceResult, error) {
	logger := utils.SugarContext(ctx)

	if status == "" {
		status = "on"
	}

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
		return common.ServiceResult{}, err
	}

	// 查询完整记录（包含默认值和时间戳）
	detail, err := dao.SPictureBook.Where(dao.SPictureBook.BookId.Eq(bookData.BookId)).First()
	if err != nil {
		logger.Errorw("PictureBookService Create First error", "error", err)
		return common.ServiceResult{}, err
	}

	return common.NewServiceResult(toPictureBookItem(detail)), nil
}
