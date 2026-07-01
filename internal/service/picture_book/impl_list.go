package picture_book

import (
	"context"

	"tool-agent/internal/common"
	"tool-agent/internal/dao"
	"tool-agent/model"
	"tool-agent/utils"

	"gorm.io/gen"
)

// ListResponseData 绘本列表响应数据
type ListResponseData struct {
	List   []PictureBookItem `json:"list"`
	Count  int64             `json:"count"`
	Offset int               `json:"offset"`
	Limit  int               `json:"limit"`
}

// PictureBookItem 绘本信息
type PictureBookItem struct {
	BookId     string `json:"book_id"`
	Title      string `json:"title"`
	Icon       string `json:"icon"`
	CategoryId string `json:"category_id"`
	Type       int    `json:"type"`
	Status     string `json:"status"`
	Position   int    `json:"position"`
}

// toPictureBookItem 将模型转换为响应项
func toPictureBookItem(m *model.SPictureBook) PictureBookItem {
	return PictureBookItem{
		BookId:     m.BookId,
		Title:      m.Title,
		Icon:       m.Icon,
		CategoryId: m.CategoryId,
		Type:       m.Type,
		Status:     m.Status,
		Position:   m.Position,
	}
}

// List 查询绘本列表
func (s *Service) List(ctx context.Context, title string, bookType int, status string, offset, limit int) (common.ServiceResult, error) {
	var (
		logger = utils.SugarContext(ctx)
		result = common.NewServiceResult()
	)
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	book := dao.SPictureBook
	where := []gen.Condition{}
	if title != "" {
		where = append(where, book.Title.Like("%"+title+"%"))
	}
	if bookType > 0 {
		where = append(where, book.Type.Eq(bookType))
	}
	if status != "" {
		where = append(where, book.Status.Eq(status))
	}

	list, count, err := book.Where(where...).Order(book.Id.Desc()).FindByPage(offset, limit)
	if err != nil {
		logger.Errorw("PictureBookService List FindByPage error", "error", err)
		return result, err
	}

	items := make([]PictureBookItem, 0, len(list))
	for _, m := range list {
		items = append(items, toPictureBookItem(m))
	}

	result.Data = ListResponseData{
		List:   items,
		Count:  count,
		Offset: offset,
		Limit:  limit,
	}
	result.SetMessage("操作成功")
	return result, nil
}
