package service

import (
	"fmt"

	appDTO "github.com/muharihar/d3ta-go/modules/email/la/application/dto"
	domRepo "github.com/muharihar/d3ta-go/modules/email/la/domain/repository"
	domSchema "github.com/muharihar/d3ta-go/modules/email/la/domain/schema"
	domSchemaET "github.com/muharihar/d3ta-go/modules/email/la/domain/schema/email_template"
	infRepo "github.com/muharihar/d3ta-go/modules/email/la/infrastructure/repository"
	sysError "github.com/muharihar/d3ta-go/system/error"
	"github.com/muharihar/d3ta-go/system/handler"
	"github.com/muharihar/d3ta-go/system/identity"
)

// NewEmailService new EmailService
func NewEmailService(h *handler.Handler) (*EmailService, error) {
	var err error

	svc := new(EmailService)
	svc.handler = h

	if svc.repoEmailTpl, err = infRepo.NewEmailTemplateRepo(h); err != nil {
		return nil, err
	}
	if svc.repoEmail, err = infRepo.NewEmailRepo(h); err != nil {
		return nil, err
	}

	return svc, nil
}

// EmailService type
type EmailService struct {
	BaseService
	repoEmail    domRepo.IEmailRepo
	repoEmailTpl domRepo.IEmailTemplateRepo
}

// Send send Email
func (s *EmailService) Send(req *appDTO.SendEmailReqDTO, i identity.Identity) (*appDTO.SendEmailResDTO, error) {
	// authorization
	if (i.CanAccessCurrentRequest() == false) && (i.CanAccess("", "system.module.email.send", "EXECUTE", nil) == false) {
		errMsg := fmt.Sprintf("You are not authorized to access [`%s.%s`]",
			i.RequestInfo.RequestObject, i.RequestInfo.RequestAction)
		return nil, sysError.CustomForbiddenAccess(errMsg)
	}

	// request domain
	reqDom := domSchema.SendEmailRequest{
		TemplateCode: req.TemplateCode,
		From:         req.From,
		To:           req.To,
		CC:           req.CC,
		BCC:          req.BCC,
		TemplateData: req.TemplateData,
	}

	if err := reqDom.Validate(); err != nil {
		return nil, err
	}

	// retrieve and assign email template
	// -->
	reqET := domSchemaET.ETFindByCodeRequest{
		Code: req.TemplateCode,
	}
	tpl, err := s.repoEmailTpl.FindByCode(&reqET, i)
	if err != nil {
		return nil, err
	}
	reqDom.Template = &tpl.Data
	// <--

	res, err := s.repoEmail.Send(&reqDom, i)
	if err != nil {
		return nil, err
	}

	// response - dto
	resDTO := new(appDTO.SendEmailResDTO)
	resDTO.TemplateCode = res.TemplateCode
	resDTO.Status = res.Status

	return resDTO, nil
}
