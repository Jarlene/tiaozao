package errors

// 业务错误码
// 0 = 成功
// 400xx = 请求参数/验证错误
// 401xx = 认证/鉴权错误
// 403xx = 权限错误
// 404xx = 资源不存在
// 500xx = 服务端错误

const (
	// 通用
	Success         = 0
	ErrBadRequest   = 40001
	ErrEmailExists  = 40002
	ErrWeakPassword = 40003
	ErrTooManyFiles = 40004 // 图片数量超出限制
	ErrFileTooLarge = 40005 // 文件大小超出限制
	ErrFileFormat   = 40006 // 文件格式不支持
	ErrCategoryHasChildren = 40007 // 分类下有子分类，无法删除
	ErrCategoryHasProducts = 40008 // 分类下有商品，无法删除

	// 认证
	ErrInvalidCredentials = 40101
	ErrInvalidToken       = 40102
	ErrInvalidRefresh     = 40103

	// 权限
	ErrForbidden = 40301 // 无权限操作
	ErrCannotReviewOwn = 40302 // 不能评价自己的商品

	// 资源
	ErrProductNotFound  = 40401
	ErrImageNotFound    = 40402
	ErrCategoryNotFound = 40403
	ErrReviewNotFound   = 40404

	// 订单
	ErrOrderNotFound = 40405

	// 服务端
	ErrInternal = 50001
	// 评论
	ErrAlreadyReviewed  = 40009 // 已评价过该商品
	ErrAlreadyReplied   = 40010 // 已回复过该评论
	ErrReviewContentTooLong = 40011 // 评论内容超长
	ErrInvalidRating    = 40012 // 评分无效

	// 库存
	ErrInsufficientStock = 40013 // 库存不足
)

var MessageMap = map[int]string{
	Success:               "ok",
	ErrBadRequest:         "请求参数无效",
	ErrEmailExists:        "邮箱已被注册",
	ErrWeakPassword:       "密码不符合策略（最少8位，需包含大小写字母和数字）",
	ErrTooManyFiles:       "图片数量超出限制（最多9张）",
	ErrFileTooLarge:       "文件大小超出限制（最大5MB）",
	ErrFileFormat:         "文件格式不支持（仅支持jpg/png/webp）",
	ErrInvalidCredentials: "邮箱或密码错误",
	ErrInvalidToken:       "令牌无效或已过期",
	ErrInvalidRefresh:     "Refresh Token 无效",
	ErrForbidden:          "无权进行此操作",
	ErrProductNotFound:    "商品不存在",
	ErrImageNotFound:      "图片不存在",
	ErrCategoryNotFound:   "分类不存在",
	ErrCategoryHasChildren: "该分类下存在子分类，无法删除",
	ErrCategoryHasProducts: "该分类下存在商品，无法删除",
	ErrInternal:           "服务器内部错误",
	ErrAlreadyReviewed:    "您已评价过该商品",
	ErrAlreadyReplied:     "已回复过该评论",
	ErrReviewContentTooLong: "评论内容超出长度限制（最多1000字）",
	ErrInvalidRating:      "评分无效（1-5星）",
	ErrCannotReviewOwn:    "不能评价自己的商品",
	ErrReviewNotFound:     "评论不存在",
	ErrOrderNotFound:      "订单不存在",
	ErrInsufficientStock:  "库存不足",
}
