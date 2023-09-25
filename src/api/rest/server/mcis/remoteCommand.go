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
	"net/http"

	"github.com/cloud-barista/cb-tumblebug/src/core/common"
	"github.com/cloud-barista/cb-tumblebug/src/core/mcis"
	"github.com/labstack/echo/v4"
)

type RestPostCmdMcisVmResponse struct {
	Result string `json:"result"`
}

type RestPostCmdMcisResponse struct {
	McisId string `json:"mcisId"`
	VmId   string `json:"vmId"`
	VmIp   string `json:"vmIp"`
	Result string `json:"result"`
}

type RestPostCmdMcisResponseWrapper struct {
	ResultArray []RestPostCmdMcisResponse `json:"resultArray"`
}

// RestPostCmdMcis godoc
// @Summary Send a command to specified MCIS
// @Description Send a command to specified MCIS
// @Tags [Infra service] MCIS Remote command
// @Accept  json
// @Produce  json
// @Param nsId path string true "Namespace ID" default(ns01)
// @Param mcisId path string true "MCIS ID" default(mcis01)
// @Param mcisCmdReq body mcis.McisCmdReq true "MCIS Command Request"
// @Param subGroupId query string false "subGroupId to apply the command only for VMs in subGroup of MCIS" default()
// @Param vmId query string false "vmId to apply the command only for a VM in MCIS" default()
// @Success 200 {object} RestPostCmdMcisResponseWrapper
// @Failure 404 {object} common.SimpleMsg
// @Failure 500 {object} common.SimpleMsg
// @Router /ns/{nsId}/cmd/mcis/{mcisId} [post]
func RestPostCmdMcis(c echo.Context) error {

	nsId := c.Param("nsId")
	mcisId := c.Param("mcisId")
	subGroupId := c.QueryParam("subGroupId")
	vmId := c.QueryParam("vmId")

	req := &mcis.McisCmdReq{}
	if err := c.Bind(req); err != nil {
		return err
	}

	resultArray, err := mcis.RemoteCommandToMcis(nsId, mcisId, subGroupId, vmId, req)
	if err != nil {
		mapA := map[string]string{"message": err.Error()}
		return c.JSON(http.StatusInternalServerError, &mapA)
	}

	content := RestPostCmdMcisResponseWrapper{}

	for _, v := range resultArray {

		resultTmp := RestPostCmdMcisResponse{}
		resultTmp.McisId = mcisId
		resultTmp.VmId = v.VmId
		resultTmp.VmIp = v.VmIp
		resultTmp.Result = v.Result
		content.ResultArray = append(content.ResultArray, resultTmp)
	}

	common.PrintJsonPretty(content)

	return c.JSON(http.StatusOK, content)

}
