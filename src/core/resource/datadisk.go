/*
Copyright 2019 The Cloud-Barista Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package resource is to manage multi-cloud infra resource
package resource

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"

	"github.com/cloud-barista/cb-tumblebug/src/core/common"
	"github.com/cloud-barista/cb-tumblebug/src/core/common/label"
	"github.com/cloud-barista/cb-tumblebug/src/core/model"
	"github.com/cloud-barista/cb-tumblebug/src/kvstore/kvstore"

	validator "github.com/go-playground/validator/v10"
)

// TbDataDiskReqStructLevelValidation func is for Validation
func TbDataDiskReqStructLevelValidation(sl validator.StructLevel) {

	u := sl.Current().Interface().(model.TbDataDiskReq)

	err := common.CheckString(u.Name)
	if err != nil {
		// ReportError(field interface{}, fieldName, structFieldName, tag, param string)
		sl.ReportError(u.Name, "name", "Name", err.Error(), "")
	}
}

// CreateDataDisk accepts DataDisk creation request, creates and returns an TB dataDisk object
func CreateDataDisk(nsId string, u *model.TbDataDiskReq, option string) (model.TbDataDiskInfo, error) {

	resourceType := model.StrDataDisk

	err := common.CheckString(nsId)
	if err != nil {
		log.Error().Err(err).Msg("")
		return model.TbDataDiskInfo{}, err
	}

	if option != "register" { // fields validation
		err = validate.Struct(u)
		if err != nil {
			if _, ok := err.(*validator.InvalidValidationError); ok {
				log.Err(err).Msg("")
				return model.TbDataDiskInfo{}, err
			}

			return model.TbDataDiskInfo{}, err
		}
	}

	check, err := CheckResource(nsId, resourceType, u.Name)

	if check {
		err := fmt.Errorf("The dataDisk %s already exists.", u.Name)
		return model.TbDataDiskInfo{}, err
	}

	if err != nil {
		err := fmt.Errorf("Failed to check the existence of the dataDisk %s.", u.Name)
		return model.TbDataDiskInfo{}, err
	}

	uid := common.GenUid()

	requestBody := model.SpiderDiskReqInfoWrapper{
		ConnectionName: u.ConnectionName,
		ReqInfo: model.SpiderDiskInfo{
			Name:     uid,
			CSPid:    u.CspResourceId, // for option=register
			DiskType: u.DiskType,
			DiskSize: u.DiskSize,
		},
	}

	var tempSpiderDiskInfo *model.SpiderDiskInfo

	client := resty.New().SetCloseConnection(true)
	client.SetAllowGetMethodPayload(true)

	req := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestBody).
		SetResult(&model.SpiderDiskInfo{}) // or SetResult(AuthSuccess{}).
		//SetError(&AuthError{}).       // or SetError(AuthError{}).

	var resp *resty.Response
	// var err error

	var url string
	if option == "register" && u.CspResourceId == "" {
		url = fmt.Sprintf("%s/disk/%s", model.SpiderRestUrl, u.Name)
		resp, err = req.Get(url)
	} else if option == "register" && u.CspResourceId != "" {
		url = fmt.Sprintf("%s/regdisk", model.SpiderRestUrl)
		resp, err = req.Post(url)
	} else { // option != "register"
		url = fmt.Sprintf("%s/disk", model.SpiderRestUrl)
		resp, err = req.Post(url)
	}

	if err != nil {
		log.Error().Err(err).Msg("")
		err := fmt.Errorf("an error occurred while requesting to CB-Spider")
		return model.TbDataDiskInfo{}, err
	}

	fmt.Printf("HTTP Status code: %d \n", resp.StatusCode())
	switch {
	case resp.StatusCode() >= 400 || resp.StatusCode() < 200:
		err := fmt.Errorf(string(resp.Body()))
		fmt.Println("body: ", string(resp.Body()))
		log.Error().Err(err).Msg("")
		return model.TbDataDiskInfo{}, err
	}

	tempSpiderDiskInfo = resp.Result().(*model.SpiderDiskInfo)

	content := model.TbDataDiskInfo{
		ResourceType:         resourceType,
		Id:                   u.Name,
		Name:                 u.Name,
		Uid:                  uid,
		ConnectionName:       u.ConnectionName,
		DiskType:             tempSpiderDiskInfo.DiskType,
		DiskSize:             tempSpiderDiskInfo.DiskSize,
		CspResourceId:        tempSpiderDiskInfo.IId.SystemId,
		CspResourceName:      tempSpiderDiskInfo.IId.NameId,
		Status:               tempSpiderDiskInfo.Status,
		AssociatedObjectList: []string{},
		CreatedTime:          tempSpiderDiskInfo.CreatedTime,
		KeyValueList:         tempSpiderDiskInfo.KeyValueList,
		Description:          u.Description,
		IsAutoGenerated:      false,
	}
	content.ConnectionConfig, err = common.GetConnConfig(content.ConnectionName)
	if err != nil {
		err = fmt.Errorf("Cannot retrieve ConnectionConfig" + err.Error())
		log.Error().Err(err).Msg("")
	}

	if option == "register" {
		if u.CspResourceId == "" {
			content.SystemLabel = "Registered from CB-Spider resource"
		} else if u.CspResourceId != "" {
			content.SystemLabel = "Registered from CSP resource"
		}
	}

	log.Info().Msg("PUT CreateDataDisk")
	Key := common.GenResourceKey(nsId, resourceType, content.Id)
	Val, _ := json.Marshal(content)
	err = kvstore.Put(Key, string(Val))
	if err != nil {
		log.Error().Err(err).Msg("")
		return content, err
	}

	// Store label info using CreateOrUpdateLabel
	labels := map[string]string{
		model.LabelManager:         model.StrManager,
		model.LabelNamespace:       nsId,
		model.LabelLabelType:       model.StrDataDisk,
		model.LabelId:              content.Id,
		model.LabelName:            content.Name,
		model.LabelUid:             content.Uid,
		model.LabelDiskType:        content.DiskType,
		model.LabelDiskSize:        content.DiskSize,
		model.LabelCspResourceId:   content.CspResourceId,
		model.LabelCspResourceName: content.CspResourceName,
		model.LabelDescription:     content.Description,
		model.LabelCreatedTime:     content.CreatedTime.String(),
		model.LabelConnectionName:  content.ConnectionName,
	}
	err = label.CreateOrUpdateLabel(model.StrDataDisk, uid, Key, labels)
	if err != nil {
		log.Error().Err(err).Msg("")
		return content, err
	}

	return content, nil
}

// TbDataDiskUpsizeReq is a struct to handle 'Upsize dataDisk' request toward CB-Tumblebug.
type TbDataDiskUpsizeReq struct {
	DiskSize    string `json:"diskSize" validate:"required"`
	Description string `json:"description"`
}

// UpsizeDataDisk accepts DataDisk upsize request, creates and returns an TB dataDisk object
func UpsizeDataDisk(nsId string, resourceId string, u *model.TbDataDiskUpsizeReq) (model.TbDataDiskInfo, error) {

	resourceType := model.StrDataDisk

	err := common.CheckString(nsId)
	if err != nil {
		log.Error().Err(err).Msg("")
		return model.TbDataDiskInfo{}, err
	}

	err = validate.Struct(u)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			log.Err(err).Msg("")
			return model.TbDataDiskInfo{}, err
		}

		return model.TbDataDiskInfo{}, err
	}

	check, err := CheckResource(nsId, resourceType, resourceId)

	if !check {
		err := fmt.Errorf("The dataDisk %s does not exist.", resourceId)
		return model.TbDataDiskInfo{}, err
	}

	if err != nil {
		err := fmt.Errorf("Failed to check the existence of the dataDisk %s.", resourceId)
		return model.TbDataDiskInfo{}, err
	}

	dataDiskInterface, err := GetResource(nsId, resourceType, resourceId)
	if err != nil {
		err := fmt.Errorf("Failed to get the dataDisk object %s.", resourceId)
		return model.TbDataDiskInfo{}, err
	}

	dataDisk := dataDiskInterface.(model.TbDataDiskInfo)

	diskSize_as_is, _ := strconv.Atoi(dataDisk.DiskSize)
	diskSize_to_be, err := strconv.Atoi(u.DiskSize)
	if err != nil {
		err := fmt.Errorf("Failed to convert the desired disk size (%s) into int.", u.DiskSize)
		return model.TbDataDiskInfo{}, err
	}

	if !(diskSize_as_is < diskSize_to_be) {
		err := fmt.Errorf("Desired disk size (%s GB) should be > %s GB.", u.DiskSize, dataDisk.DiskSize)
		return model.TbDataDiskInfo{}, err
	}

	requestBody := model.SpiderDiskUpsizeReqWrapper{
		ConnectionName: dataDisk.ConnectionName,
		ReqInfo: model.SpiderDiskUpsizeReq{
			Size: u.DiskSize,
		},
	}

	client := resty.New().SetCloseConnection(true)
	client.SetAllowGetMethodPayload(true)

	req := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestBody)
		// SetResult(&SpiderDiskInfo{}) // or SetResult(AuthSuccess{}).
		//SetError(&AuthError{}).       // or SetError(AuthError{}).

	var resp *resty.Response
	// var err error

	url := fmt.Sprintf("%s/disk/%s/size", model.SpiderRestUrl, dataDisk.CspResourceName)
	resp, err = req.Put(url)

	if err != nil {
		log.Error().Err(err).Msg("")
		err := fmt.Errorf("an error occurred while requesting to CB-Spider")
		return model.TbDataDiskInfo{}, err
	}

	fmt.Printf("HTTP Status code: %d \n", resp.StatusCode())
	switch {
	case resp.StatusCode() >= 400 || resp.StatusCode() < 200:
		err := fmt.Errorf(string(resp.Body()))
		fmt.Println("body: ", string(resp.Body()))
		log.Error().Err(err).Msg("")
		return model.TbDataDiskInfo{}, err
	}

	/*
		isSuccessful := resp.Result().(bool)
		if isSuccessful == false {
			err := fmt.Errorf("Failed to upsize the dataDisk %s", resourceId)
			return model.TbDataDiskInfo{}, err
		}
	*/

	content := dataDisk
	content.DiskSize = u.DiskSize
	content.Description = u.Description

	log.Info().Msg("PUT UpsizeDataDisk")
	Key := common.GenResourceKey(nsId, resourceType, content.Id)
	Val, _ := json.Marshal(content)
	err = kvstore.Put(Key, string(Val))
	if err != nil {
		log.Error().Err(err).Msg("")
		return content, err
	}
	return content, nil
}
