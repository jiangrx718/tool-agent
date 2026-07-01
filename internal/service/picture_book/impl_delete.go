package picture_book

import (
	"context"

	"tool-agent/internal/common"
	"tool-agent/internal/dao"
	"tool-agent/utils"
)

// Delete 删除绘本
func (s *Service) Delete(ctx context.Context, bookId string) (common.ServiceResult, error) {
	var (
		logger = utils.SugarContext(ctx)
		result = common.NewServiceResult()
	)

	book := dao.SPictureBook

	// 校验绘本是否存在
	count, err := book.Where(book.BookId.Eq(bookId)).Count()
	if err != nil {
		logger.Errorw("PictureBookService Delete Count error", "book_id", bookId, "error", err)
		return result, err
	}
	if count == 0 {
		result.SetError(&common.ServiceError{Code: 400, Message: "绘本不存在"})
		return result, nil
	}

	if _, err := book.Where(book.BookId.Eq(bookId)).Delete(); err != nil {
		logger.Errorw("PictureBookService Delete error", "book_id", bookId, "error", err)
		return result, err
	}

	result.SetMessage("操作成功")
	return result, nil
}
