package model

type SPictureBook struct {
	BaseModelFieldId
	BookId     string `gorm:"column:book_id;type:char(36);comment:绘本id;NOT NULL" json:"book_id"`
	Title      string `gorm:"column:title;type:varchar(1024);comment:绘本标题;NOT NULL" json:"title"`
	Icon       string `gorm:"column:icon;type:varchar(1024);comment:绘本封面;NOT NULL" json:"icon"`
	CategoryId string `gorm:"column:category_id;type:char(36);comment:绘本所属栏目;NOT NULL" json:"category_id"`
	Type       int    `gorm:"column:type;type:int(11);default:0;comment:1中文绘本,2英文绘本,3古诗绘本,4英语词汇;NOT NULL" json:"type"`
	Status     string `gorm:"column:status;type:varchar(20);default:on;comment:状态,on启用,off禁用;NOT NULL" json:"status"`
	Position   int    `gorm:"column:position;type:int(11);default:0;comment:排序位置;NOT NULL" json:"position"`
	BaseModelFieldTime
}

func (m *SPictureBook) TableName() string {
	return "s_picture_book"
}
