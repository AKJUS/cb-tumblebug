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

// Package mcis is to handle REST API for mcis
package mcis

import (
	"fmt"
	"net/http"

	"github.com/cloud-barista/cb-tumblebug/src/core/common"
	"github.com/cloud-barista/cb-tumblebug/src/core/mcis"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// RestPostMcisPolicy godoc
// @ID PostMcisPolicy
// @Summary Create MCIS Automation policy
// @Description Create MCIS Automation policy
// @Tags [Infra service] MCIS Auto control policy management (WIP)
// @Accept  json
// @Produce  json
// @Param nsId path string true "Namespace ID" default(ns01)
// @Param mcisId path string true "MCIS ID" default(mcis01)
// @Param mcisPolicyReq body mcis.McisPolicyReq true "Details for an MCIS automation policy request"
// @Success 200 {object} mcis.McisPolicyInfo
// @Failure 404 {object} common.SimpleMsg
// @Failure 500 {object} common.SimpleMsg
// @Router /ns/{nsId}/policy/mcis/{mcisId} [post]
func RestPostMcisPolicy(c echo.Context) error {
	reqID, idErr := common.StartRequestWithLog(c)
	if idErr != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": idErr.Error()})
	}
	nsId := c.Param("nsId")
	mcisId := c.Param("mcisId")

	req := &mcis.McisPolicyReq{}
	if err := c.Bind(req); err != nil {
		return common.EndRequestWithLog(c, reqID, err, nil)
	}

	content, err := mcis.CreateMcisPolicy(nsId, mcisId, req)
	return common.EndRequestWithLog(c, reqID, err, content)
}

// RestGetMcisPolicy godoc
// @ID GetMcisPolicy
// @Summary Get MCIS Policy
// @Description Get MCIS Policy
// @Tags [Infra service] MCIS Auto control policy management (WIP)
// @Accept  json
// @Produce  json
// @Param nsId path string true "Namespace ID" default(ns01)
// @Param mcisId path string true "MCIS ID" default(mcis01)
// @Success 200 {object} mcis.McisPolicyInfo
// @Failure 404 {object} common.SimpleMsg
// @Failure 500 {object} common.SimpleMsg
// @Router /ns/{nsId}/policy/mcis/{mcisId} [get]
func RestGetMcisPolicy(c echo.Context) error {
	reqID, idErr := common.StartRequestWithLog(c)
	if idErr != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": idErr.Error()})
	}

	nsId := c.Param("nsId")
	mcisId := c.Param("mcisId")

	result, err := mcis.GetMcisPolicyObject(nsId, mcisId)
	if err != nil {
		errorMessage := fmt.Errorf("Error to find McisPolicyObject : " + mcisId + "ERROR : " + err.Error())
		return common.EndRequestWithLog(c, reqID, errorMessage, nil)
	}

	if result.Id == "" {
		errorMessage := fmt.Errorf("Failed to find McisPolicyObject : " + mcisId)
		return common.EndRequestWithLog(c, reqID, errorMessage, nil)
	}
	return common.EndRequestWithLog(c, reqID, err, result)
}

// Response structure for RestGetAllMcisPolicy
type RestGetAllMcisPolicyResponse struct {
	McisPolicy []mcis.McisPolicyInfo `json:"mcisPolicy"`
}

// RestGetAllMcisPolicy godoc
// @ID GetAllMcisPolicy
// @Summary List all MCIS policies
// @Description List all MCIS policies
// @Tags [Infra service] MCIS Auto control policy management (WIP)
// @Accept  json
// @Produce  json
// @Param nsId path string true "Namespace ID" default(ns01)
// @Success 200 {object} RestGetAllMcisPolicyResponse
// @Failure 404 {object} common.SimpleMsg
// @Failure 500 {object} common.SimpleMsg
// @Router /ns/{nsId}/policy/mcis [get]
func RestGetAllMcisPolicy(c echo.Context) error {
	reqID, idErr := common.StartRequestWithLog(c)
	if idErr != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": idErr.Error()})
	}
	nsId := c.Param("nsId")
	log.Debug().Msg("[Get MCIS Policy List]")

	result, err := mcis.GetAllMcisPolicyObject(nsId)
	if err != nil {
		return common.EndRequestWithLog(c, reqID, err, nil)
	}

	content := RestGetAllMcisPolicyResponse{}
	content.McisPolicy = result
	return common.EndRequestWithLog(c, reqID, err, content)
}

/*
	function RestPutMcisPolicy not yet implemented

// RestPutMcisPolicy godoc
// @ID PutMcisPolicy
// @Summary Update MCIS Policy
// @Description Update MCIS Policy
// @Tags [Infra service] MCIS Auto control policy management (WIP)
// @Accept  json
// @Produce  json
// @Param mcisInfo body McisPolicyInfo true "Details for an MCIS Policy object"
// @Success 200 {object} McisPolicyInfo
// @Failure 404 {object} common.SimpleMsg
// @Failure 500 {object} common.SimpleMsg
// @Router /ns/{nsId}/policy/mcis/{mcisId} [put]
*/
func RestPutMcisPolicy(c echo.Context) error {
	return nil
}

// DelMcisPolicy godoc
// @ID DelMcisPolicy
// @Summary Delete MCIS Policy
// @Description Delete MCIS Policy
// @Tags [Infra service] MCIS Auto control policy management (WIP)
// @Accept  json
// @Produce  json
// @Param nsId path string true "Namespace ID" default(ns01)
// @Param mcisId path string true "MCIS ID" default(mcis01)
// @Success 200 {object} common.SimpleMsg
// @Failure 404 {object} common.SimpleMsg
// @Router /ns/{nsId}/policy/mcis/{mcisId} [delete]
func RestDelMcisPolicy(c echo.Context) error {
	reqID, idErr := common.StartRequestWithLog(c)
	if idErr != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": idErr.Error()})
	}
	nsId := c.Param("nsId")
	mcisId := c.Param("mcisId")

	err := mcis.DelMcisPolicy(nsId, mcisId)
	result := map[string]string{"message": "Deleted the MCIS Policy info"}
	return common.EndRequestWithLog(c, reqID, err, result)
}

// RestDelAllMcisPolicy godoc
// @ID DelAllMcisPolicy
// @Summary Delete all MCIS policies
// @Description Delete all MCIS policies
// @Tags [Infra service] MCIS Auto control policy management (WIP)
// @Accept  json
// @Produce  json
// @Param nsId path string true "Namespace ID" default(ns01)
// @Success 200 {object} common.SimpleMsg
// @Failure 404 {object} common.SimpleMsg
// @Router /ns/{nsId}/policy/mcis [delete]
func RestDelAllMcisPolicy(c echo.Context) error {
	reqID, idErr := common.StartRequestWithLog(c)
	if idErr != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": idErr.Error()})
	}
	nsId := c.Param("nsId")
	result, err := mcis.DelAllMcisPolicy(nsId)
	return common.EndRequestWithLog(c, reqID, err, result)
}
