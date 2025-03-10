package services

import (
	"bytes"
	"context"
	"field-service/common/gcs"
	"field-service/common/util"
	errConstant "field-service/constants/error"
	"field-service/domain/dto"
	"field-service/domain/models"
	"field-service/repositories"
	"fmt"
	"io"
	"mime/multipart"
	"path"
	"time"
)

type FieldService struct {
	repository repositories.IRepositoryRegistry
	gcs        gcs.IGCSClient
}

type IFieldService interface {
	GetAllWithPagination(context.Context, *dto.FieldRequestParam) (*util.PaginationResult, error)
	GetAllWithoutPagination(context.Context) ([]dto.FieldResponse, error)
	GetByUUID(context.Context, string) (*dto.FieldResponse, error)
	Create(context.Context, *dto.FieldRequest) (*dto.FieldResponse, error)
	Update(context.Context, string, *dto.UpdateFieldRequest) (*dto.FieldResponse, error)
	Delete(context.Context, string) error
}

func NewFieldService(repository repositories.IRepositoryRegistry, gcs gcs.IGCSClient) IFieldService {
	return &FieldService{
		repository: repository,
		gcs:        gcs,
	}
}

func (s *FieldService) GetAllWithPagination(ctx context.Context, param *dto.FieldRequestParam) (*util.PaginationResult, error) {
	fields, total, err := s.repository.GetField().FindAllWithPagination(ctx, param)
	if err != nil {
		return nil, err
	}

	fieldResults := make([]*dto.FieldResponse, 0, len(fields))
	for _, field := range fields {
		fieldResults = append(fieldResults, &dto.FieldResponse{
			UUID:         field.UUID,
			Code:         field.Code,
			Name:         field.Name,
			PricePerHour: field.PricePerHour,
			Images:       field.Images,
			CreatedAt:    field.CreatedAt,
			UpdatedAt:    field.UpdatedAt,
		})
	}

	pagination := &util.PaginationParam{
		Count: total,
		Page:  param.Page,
		Limit: param.Limit,
		Data:  fieldResults,
	}

	response := util.GeneratePagination(*pagination)
	return &response, nil
}

func (s *FieldService) GetAllWithoutPagination(ctx context.Context) ([]dto.FieldResponse, error) {
	fields, err := s.repository.GetField().FindAllWithoutPagination(ctx)
	if err != nil {
		return nil, err
	}

	fieldResults := make([]dto.FieldResponse, 0, len(fields))
	for _, field := range fields {
		fieldResults = append(fieldResults, dto.FieldResponse{
			UUID:         field.UUID,
			Code:         field.Code,
			Name:         field.Name,
			PricePerHour: field.PricePerHour,
			Images:       field.Images,
		})
	}

	return fieldResults, nil
}

func (s *FieldService) GetByUUID(ctx context.Context, uuid string) (*dto.FieldResponse, error) {
	field, err := s.repository.GetField().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	fieldResult := dto.FieldResponse{
		UUID:         field.UUID,
		Code:         field.Code,
		Name:         field.Name,
		PricePerHour: field.PricePerHour,
		Images:       field.Images,
		CreatedAt:    field.CreatedAt,
		UpdatedAt:    field.UpdatedAt,
	}

	return &fieldResult, nil
}

func (f *FieldService) validateUpload(images []multipart.FileHeader) error {
	if images == nil || len(images) == 0 {
		return errConstant.ErrInValidUploadFile
	}

	for _, image := range images {
		if image.Size > 1024*1024*5 {
			return errConstant.ErrSizeToBig
		}
	}

	return nil
}

func (s *FieldService) processAndUploadImage(ctx context.Context, image multipart.FileHeader) (string, error) {
	file, err := image.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	buffer := new(bytes.Buffer)
	_, err = io.Copy(buffer, file)
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("images/%s-%s.%s", time.Now().Format("20060102150405"), image.Filename, path.Ext(image.Filename))
	url, err := s.gcs.UploadFile(ctx, filename, buffer.Bytes())
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *FieldService) uploadImage(ctx context.Context, images []multipart.FileHeader) ([]string, error) {
	err := s.validateUpload(images)
	if err != nil {
		return nil, err
	}

	urls := make([]string, 0, len(images))
	for _, image := range images {
		url, err := s.processAndUploadImage(ctx, image)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}

	return urls, nil
}

func (s *FieldService) Create(ctx context.Context, request *dto.FieldRequest) (*dto.FieldResponse, error) {
	imageUrl, err := s.uploadImage(ctx, request.Images)
	if err != nil {
		return nil, err
	}

	field, err := s.repository.GetField().Create(ctx, &models.Field{
		Code:         request.Code,
		Name:         request.Name,
		PricePerHour: request.PricePerHour,
		Images:       imageUrl,
	})
	if err != nil {
		return nil, err
	}

	response := &dto.FieldResponse{
		UUID:         field.UUID,
		Code:         field.Code,
		Name:         field.Name,
		PricePerHour: field.PricePerHour,
		Images:       field.Images,
		CreatedAt:    field.CreatedAt,
		UpdatedAt:    field.UpdatedAt,
	}
	return response, nil
}

func (s *FieldService) Update(ctx context.Context, uuid string, request *dto.UpdateFieldRequest) (*dto.FieldResponse, error) {
	field, err := s.repository.GetField().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	var imageUrl []string
	if request.Images == nil {
		imageUrl = field.Images
	} else {
		imageUrl, err = s.uploadImage(ctx, request.Images)
		if err != nil {
			return nil, err
		}
	}

	fieldResult, err := s.repository.GetField().Update(ctx, uuid, &models.Field{
		Code:         request.Code,
		Name:         request.Name,
		PricePerHour: request.PricePerHour,
		Images:       imageUrl,
	})
	if err != nil {
		return nil, err
	}
	return &dto.FieldResponse{
		UUID:         fieldResult.UUID,
		Code:         fieldResult.Code,
		Name:         fieldResult.Name,
		PricePerHour: fieldResult.PricePerHour,
		Images:       fieldResult.Images,
		CreatedAt:    fieldResult.CreatedAt,
		UpdatedAt:    fieldResult.UpdatedAt,
	}, nil
}

func (s *FieldService) Delete(ctx context.Context, uuid string) error {
	_, err := s.repository.GetField().FindByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	err = s.repository.GetField().Delete(ctx, uuid)
	if err != nil {
		return err
	}

	return nil

}
