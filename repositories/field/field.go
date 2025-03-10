package repositories

import (
	"context"
	"errors"
	"field-service/domain/dto"
	"field-service/domain/models"
	"fmt"

	errWrap "field-service/common/error"
	errConstant "field-service/constants/error"
	errField "field-service/constants/error/field"

	"gorm.io/gorm"
)

type FieldRepository struct {
	db *gorm.DB
}

type IFieldRepository interface {
	FindAllWithPagination(context.Context, *dto.FieldRequestParam) ([]models.Field, int64, error)
	FindAllWithoutPagination(context.Context) ([]models.Field, error)
	FindByUUID(context.Context, string) (*models.Field, error)
	Create(context.Context, *models.Field) (*models.Field, error)
	Update(context.Context, string, *models.Field) (*models.Field, error)
	Delete(context.Context, string) error
}

func NewFieldRepository(db *gorm.DB) IFieldRepository {
	return &FieldRepository{db: db}
}

func (r *FieldRepository) FindAllWithPagination(ctx context.Context, param *dto.FieldRequestParam) ([]models.Field, int64, error) {
	var (
		fields []models.Field
		sort   string
		total  int64
	)

	if param.SortColumn != nil {
		sort = fmt.Sprintf("%s %s", *param.SortColumn, *param.SortOrder)
	} else {
		sort = "created_at desc"
	}

	limit := param.Limit
	offset := (param.Page - 1) * limit
	err := r.db.
		WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order(sort).
		Find(&fields).
		Error

	if err != nil {
		return nil, 0, errWrap.WrapError(errConstant.ErrSQLError)
	}

	err = r.db.
		WithContext(ctx).
		Model(&fields).
		Count(&total).
		Error
	if err != nil {
		return nil, 0, errWrap.WrapError(errConstant.ErrSQLError)
	}

	return fields, total, nil
}

func (r *FieldRepository) FindAllWithoutPagination(ctx context.Context) ([]models.Field, error) {
	var fields []models.Field
	err := r.db.
		WithContext(ctx).
		Find(&fields).
		Error
	if err != nil {
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return fields, nil
}

func (r *FieldRepository) FindByUUID(ctx context.Context, uuid string) (*models.Field, error) {
	var field models.Field
	err := r.db.
		WithContext(ctx).
		Where("uuid = ?", uuid).
		First(&field).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errWrap.WrapError(errField.ErrFieldNotFound)
		}
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return &field, nil
}

func (r *FieldRepository) Create(ctx context.Context, req *models.Field) (*models.Field, error) {
	field := models.Field{
		UUID:         req.UUID,
		Code:         req.Code,
		Name:         req.Name,
		Images:       req.Images,
		PricePerHour: req.PricePerHour,
	}

	err := r.db.WithContext(ctx).Create(&field).Error
	if err != nil {
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return &field, nil
}

func (r *FieldRepository) Update(ctx context.Context, uuid string, req *models.Field) (*models.Field, error) {
	field := models.Field{
		Code:         req.Code,
		Name:         req.Name,
		Images:       req.Images,
		PricePerHour: req.PricePerHour,
	}

	err := r.db.WithContext(ctx).Where("uuid = ?", uuid).Updates(&field).Error
	if err != nil {
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return &field, nil
}

func (f *FieldRepository) Delete(ctx context.Context, uuid string) error {
	err := f.db.WithContext(ctx).Where("uuid = ?", uuid).Delete(&models.Field{}).Error
	if err != nil {
		return errWrap.WrapError(errConstant.ErrSQLError)
	}
	return nil
}
