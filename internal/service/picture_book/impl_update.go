package picture_book

import (
	"context"

	"tool-agent/internal/common"
	"tool-agent/internal/dao"
	"tool-agent/utils"
)

// Update 更新绘本
func (s *Service) Update(ctx context.Context, bookId, title, icon, categoryId string, bookType int, status string, position int) (common.ServiceResult, error) {
	var (
		logger = utils.SugarContext(ctx)
		result = common.NewServiceResult()
	)

	book := dao.SPictureBook

	// 校验绘本是否存在
	count, err := book.Where(book.BookId.Eq(bookId)).Count()
	if err != nil {
		logger.Errorw("PictureBookService Update Count error", "book_id", bookId, "error", err)
		return result, err
	}
	if count == 0 {
		result.SetError(&common.ServiceError{Code: 400, Message: "绘本不存在"})
		return result, nil
	}

	updates := map[string]interface{}{
		"title":       title,
		"icon":        icon,
		"category_id": categoryId,
		"type":        bookType,
		"status":      status,
		"position":    position,
	}

	if _, err := book.Where(book.BookId.Eq(bookId)).Updates(updates); err != nil {
		logger.Errorw("PictureBookService Update Updates error", "book_id", bookId, "error", err)
		return result, err
	}

	// 查询更新后的记录
	detail, err := book.Where(book.BookId.Eq(bookId)).First()
	if err != nil {
		logger.Errorw("PictureBookService Update First error", "book_id", bookId, "error", err)
		return result, err
	}

	result.Data = toPictureBookItem(detail)
	result.SetMessage("操作成功")
	return result, nil
}
