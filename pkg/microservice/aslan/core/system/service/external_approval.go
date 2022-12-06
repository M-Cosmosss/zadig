/*
 * Copyright 2022 The KodeRover Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	commonmodels "github.com/koderover/zadig/pkg/microservice/aslan/core/common/repository/models"
	"github.com/koderover/zadig/pkg/microservice/aslan/core/common/repository/mongodb"
	"github.com/koderover/zadig/pkg/setting"
	e "github.com/koderover/zadig/pkg/tool/errors"
	"github.com/koderover/zadig/pkg/tool/lark"
)

func ListExternalApproval(_type string, log *zap.SugaredLogger) ([]*commonmodels.ExternalApproval, error) {
	resp, err := mongodb.NewExternalApprovalColl().List(context.Background(), _type)
	if err != nil {
		log.Errorf("list external approval error: %v", err)
		return nil, e.ErrListExternalApproval.AddErr(err)
	}

	return resp, nil
}

func CreateExternalApproval(args *commonmodels.ExternalApproval, log *zap.SugaredLogger) error {
	oid, err := mongodb.NewExternalApprovalColl().Create(context.Background(), args)
	if err != nil {
		log.Errorf("create external approval error: %v", err)
		return e.ErrCreateExternalApproval.AddErr(err)
	}

	approvalCode, err := createLarkDefaultApprovalDefinition(lark.NewClient(args.AppID, args.AppSecret))
	if err != nil {
		return e.ErrCreateExternalApproval.AddErr(err)
	}
	args.LarkDefaultApprovalCode = approvalCode
	err = mongodb.NewExternalApprovalColl().Update(context.Background(), oid, args)
	if err != nil {
		return errors.Wrap(err, "update approval with approval code")
	}
	return nil
}

func UpdateExternalApproval(id string, args *commonmodels.ExternalApproval, log *zap.SugaredLogger) error {
	err := mongodb.NewExternalApprovalColl().Update(context.Background(), id, args)
	if err != nil {
		log.Errorf("update external approval error: %v", err)
		return e.ErrUpdateExternalApproval.AddErr(err)
	}
	return nil
}

func DeleteExternalApproval(id string, log *zap.SugaredLogger) error {
	err := mongodb.NewExternalApprovalColl().DeleteByID(context.Background(), id)
	if err != nil {
		log.Errorf("delete external approval error: %v", err)
		return e.ErrDeleteExternalApproval.AddErr(err)
	}
	return nil
}

func ValidateExternalApproval(approval *commonmodels.ExternalApproval, log *zap.SugaredLogger) error {
	switch approval.Type {
	case setting.IMLark:
		return lark.Validate(approval.AppID, approval.AppSecret)
	case setting.IMDingding:
	default:
		return e.ErrValidateExternalApproval.AddDesc("invalid type")
	}
	return nil
}

func createLarkDefaultApprovalDefinition(client *lark.Client) (string, error) {
	code, err := client.CreateApprovalDefinition(&lark.CreateApprovalDefinitionArgs{
		Name:        "Zadig 默认审批定义-OR",
		Description: "Zadig 默认审批定义-OR",
		Type:        lark.ApproveTypeOr,
	})
	return code, errors.Wrap(err, "create approval definition")
}
